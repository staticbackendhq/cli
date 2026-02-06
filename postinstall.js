"use strict";
// Thanks to author of https://github.com/sanathkr/go-npm, we were able to modify his code to work with private packages
var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) { return typeof obj; } : function (obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; };

var path = require('path'),
	mkdirp = require('mkdirp'),
	fs = require('fs');

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
var ARCH_MAPPING = {
	"ia32": "386",
	"x64": "amd64",
	"arm": "arm",
	"arm64": "arm64"
};

// Mapping between Node's `process.platform` to Golang's
var PLATFORM_MAPPING = {
	"darwin": "darwin",
	"linux": "linux",
	"win32": "windows",
	"freebsd": "freebsd"
};

async function getInstallationPath() {
	// `npm bin -g` is deprecated in npm 9+, use `npm prefix -g` instead
	const value = await execShellCommand("npm prefix -g");

	var dir = null;
	if (!value || value.length === 0) {
		// We couldn't infer path from `npm prefix`. Let's try to get it from
		// Environment variables set by NPM when it runs.
		// npm_config_prefix points to NPM's installation directory where `bin` folder is available
		// Ex: /Users/foo/.nvm/versions/node/v4.3.0
		var env = process.env;
		if (env && env.npm_config_prefix) {
			dir = path.join(env.npm_config_prefix, "bin");
		}
	} else {
		dir = path.join(value.trim(), "bin");
	}

	if (!dir) {
		throw new Error("Could not determine npm global bin directory");
	}

	// Create directory if it doesn't exist
	await mkdirp(dir);
	return dir;
}

async function verifyAndPlaceBinary(binName, binPath) {
	const sourcePath = path.join(binPath, binName);

	if (!fs.existsSync(sourcePath)) {
		throw new Error('Downloaded binary does not contain the binary specified in configuration - ' + binName);
	}

	// Get installation path for executables under node
	const installationPath = await getInstallationPath();
	const targetPath = path.join(installationPath, binName);

	// Copy the executable to the installation path
	await fs.promises.rename(sourcePath, targetPath);

	// Make sure the binary is executable
	await fs.promises.chmod(targetPath, 0o755);

	console.info("Installed cli successfully");
}

function validateConfiguration(packageJson) {

	if (!packageJson.version) {
		return "'version' property must be specified";
	}

	if (!packageJson.goBinary || _typeof(packageJson.goBinary) !== "object") {
		return "'goBinary' property must be defined and be an object";
	}

	if (!packageJson.goBinary.name) {
		return "'name' property is necessary";
	}

	if (!packageJson.goBinary.path) {
		return "'path' property is necessary";
	}
}

function parsePackageJson() {
	if (!(process.arch in ARCH_MAPPING)) {
		console.error("Installation is not supported for this architecture: " + process.arch);
		return;
	}

	if (!(process.platform in PLATFORM_MAPPING)) {
		console.error("Installation is not supported for this platform: " + process.platform);
		return;
	}

	var packageJsonPath = path.join(".", "package.json");
	if (!fs.existsSync(packageJsonPath)) {
		console.error("Unable to find package.json. " + "Please run this script at root of the package you want to be installed");
		return;
	}

	var packageJson = JSON.parse(fs.readFileSync(packageJsonPath));
	var error = validateConfiguration(packageJson);
	if (error && error.length > 0) {
		console.error("Invalid package.json: " + error);
		return;
	}

	// We have validated the config. It exists in all its glory
	var binName = packageJson.goBinary.name;
	var binPath = packageJson.goBinary.path;
	var version = packageJson.version;
	if (version[0] === 'v') version = version.substr(1); // strip the 'v' if necessary v0.0.1 => 0.0.1

	// Binary name on Windows has .exe suffix
	if (process.platform === "win32") {
		binName += ".exe";
	}


	return {
		binName: binName,
		binPath: binPath,
		version: version
	};
}

/**
 * Reads the configuration from application's package.json,
 * validates properties, copied the binary from the package and stores at
 * ./bin in the package's root. NPM already has support to install binary files
 * specific locations when invoked with "npm install -g"
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
var INVALID_INPUT = "Invalid inputs";
async function downloadBinary(url, dest) {
	const https = require('https');
	const http = require('http');

	return new Promise((resolve, reject) => {
		const protocol = url.startsWith('https') ? https : http;
		const file = fs.createWriteStream(dest);

		console.info(`Downloading ${url}`);

		protocol.get(url, (response) => {
			// Handle redirects
			if (response.statusCode === 302 || response.statusCode === 301) {
				file.close();
				fs.unlinkSync(dest);
				return downloadBinary(response.headers.location, dest).then(resolve).catch(reject);
			}

			if (response.statusCode !== 200) {
				file.close();
				fs.unlinkSync(dest);
				return reject(new Error(`Failed to download: ${response.statusCode} ${response.statusMessage}`));
			}

			response.pipe(file);

			file.on('finish', () => {
				file.close();
				resolve();
			});
		}).on('error', (err) => {
			file.close();
			fs.unlinkSync(dest);
			reject(err);
		});

		file.on('error', (err) => {
			file.close();
			fs.unlinkSync(dest);
			reject(err);
		});
	});
}

async function install(callback) {
	try {
		var opts = parsePackageJson();
		if (!opts) return callback(INVALID_INPUT);

		mkdirp.sync(opts.binPath);

		// Construct the binary filename based on platform and architecture
		const platform = PLATFORM_MAPPING[process.platform];
		const arch = ARCH_MAPPING[process.arch];
		const binaryName = `${platform}-${arch}-${opts.binName}`;

		// Get the GitHub repo URL from package.json
		const packageJsonPath = path.join(".", "package.json");
		const packageJson = JSON.parse(fs.readFileSync(packageJsonPath));

		if (!packageJson.goBinary.repo) {
			return callback(new Error("'goBinary.repo' property is required (e.g., 'owner/repo')"));
		}

		// Construct GitHub release download URL
		const downloadUrl = `https://github.com/${packageJson.goBinary.repo}/releases/download/v${opts.version}/${binaryName}`;
		const dest = path.join(opts.binPath, opts.binName);

		console.info(`Downloading binary for ${process.platform}-${process.arch}`);

		// Download the binary
		await downloadBinary(downloadUrl, dest);

		// Make it executable
		await fs.promises.chmod(dest, 0o755);

		await verifyAndPlaceBinary(opts.binName, opts.binPath);
		callback(null);
	} catch (err) {
		callback(err);
	}
}

async function uninstall(callback) {
	try {
		var opts = parsePackageJson();
		if (!opts) {
			console.info("Uninstalled cli successfully");
			return callback(null);
		}

		const installationPath = await getInstallationPath();
		const binaryPath = path.join(installationPath, opts.binName);

		// Use promises API for consistency
		await fs.promises.unlink(binaryPath);
		console.info("Uninstalled cli successfully");
		callback(null);
	} catch (ex) {
		// Ignore errors when deleting the file (might not exist)
		console.info("Uninstalled cli successfully");
		callback(null);
	}
}

// Parse command line arguments and call the right method
var actions = {
	"install": install,
	"uninstall": uninstall
};
/**
 * Executes a shell command and return it as a Promise.
 * @param cmd {string}
 * @return {Promise<string>}
 */
function execShellCommand(cmd) {
	const exec = require('child_process').exec;
	return new Promise((resolve, reject) => {
		exec(cmd, (error, stdout, stderr) => {
			if (error) {
				console.warn(error);
			}
			resolve(stdout ? stdout : stderr);
		});
	});
}

var argv = process.argv;
if (argv && argv.length > 2) {
	var cmd = process.argv[2];
	if (!actions[cmd]) {
		console.log("Invalid command. `install` and `uninstall` are the only supported commands");
		process.exit(1);
	}

	actions[cmd](function (err) {
		if (err) {
			console.error(err);
			process.exit(1);
		} else {
			process.exit(0);
		}
	});
}