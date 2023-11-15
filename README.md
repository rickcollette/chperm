
# chperm

`chperm` is a command-line utility for applying Unix permissions to folders and files. It offers functionality for both recursive and non-recursive operations and supports rollback of recent changes. Permissions are set in octal format, and an audit option is available to record changes in an Excel file.

## Features

- Apply Unix permissions to files and directories.
- Recursive and non-recursive permission setting.
- Rollback functionality to revert recent changes.
- Audit capability to track permission changes.

## Installation

### Prerequisites

- Go (for building from source)
- Git (for versioning and building)
- RPM or Debian-based Linux distribution (for package installation)

### Building from Source

Clone the repository and use the provided Makefile:

```sh
git clone https://github.com/rickcollette/chperm
cd chperm
make build
```

### Installing from Packages

Use the provided `buildpkgs.sh` script to create RPM or Debian packages:

```sh
./buildpkgs.sh create_rpm
./buildpkgs.sh create_deb
```

Then, install the package using your distribution's package manager.

## Usage

Run `chperm` with the desired options:

```sh
chperm [OPTIONS]...
```

Refer to the man page `chperm.1` for detailed usage instructions.

## Configuration

Modify the `chperm.conf` file to set default paths and permissions. The format is:

```conf
# path, permissions, recurse (bool)
# Example:
/home/${USER}, 0700,1
/home/${USER}/.ssh,0600,0
/etc/passwd*,0600,0
```

## Usage
## Description
chperm is a command-line utility for applying Unix permissions to folders and files. It can operate either recursively or non-recursively and supports rollback of the last few changes made. The permissions are specified in octal format. When used with the -audit option, it records the changes in an Excel file.
## Options

- ```-vvv``` Enable verbose output.
- ```-path``` Specify the path to apply permissions.
- ```-perms``` Specify the permissions in octal format.
- ```-recurse``` Recurse into directories.
- ```-rollback``` Rollback the last N changes.
- ```-audit``` Audit changes to permissions. This will generate an Excel file named 'audit_<timestamp>.xlsx' where <timestamp> is the time at which the application was run.
- ```-o csv``` Output audit to a csv file. (xlsx is recommended for Excel format)

## EXAMPLES

**Apply permissions recursively and audit:**

```chperm -path /path/to/folder -perms 0755 -recurse -audit```

**Apply permissions without recursion:**

```chperm -path /path/to/file -perms 0644```

**Rollback last 5 permission changes:**

```chperm -rollback 5```

## PERMISSION BITS
Permissions in Unix are represented by three groups: owner, group, and others. Each group can have read (r), write (w), and execute (x) permissions. Permissions are represented in octal format:

- **Read (r)** is 4.

- **Write (w)** is 2.

- **Execute (x)** is 1.

To combine permissions, add the values together. 

For example: read and write (rw) is 6 (4+2), and read, write, and execute (rwx) is 7 (4+2+1).

## CONFIGURATION FILE
The configuration file \fI/etc/chperm/chperm.conf\fR specifies the default paths and permissions that chperm should apply when not using command-line arguments. Each line in the file should contain a path, the permissions for that path, and a boolean flag indicating whether to recurse into subdirectories (1 for true, 0 for false). Lines beginning with a hash (#) are treated as comments and ignored.
## FILES

- **chperm.conf**: Configuration file. Found in /etc, /usr/local/etc, /etc/default, or current directory.
- **audit_<timestamp>.xlsx** Excel file generated for audit logs. <timestamp> format is 'YYYYMMDD_HHMMSS'.


## Makefile Usage for chperm

The `Makefile` for `chperm` provides various targets to automate the building and packaging of the application.

## Targets

### `all`
- Runs `directories`, `build`, and `config` targets.
- Usage: `make all`

### `directories`
- Creates the necessary directory structure for the application.
- Usage: `make directories`

### `build`
- Compiles the Go binary for `chperm`.
- Usage: `make build`

### `config`
- Copies the configuration file and man page to the appropriate directories.
- Usage: `make config`

### `clean`
- Cleans up the project by removing the `dist` and `build` directories.
- Usage: `make clean`

### `deb`
- Builds the Debian package for `chperm`.
- Depends on the `all` target.
- Usage: `make deb`

### `rpm`
- Builds the RPM package for `chperm`.
- Depends on the `all` target.
- Usage: `make rpm`

### `shar`
- Builds the SHAR (Shell Archive) package for `chperm`.
- Depends on the `all` target.
- Usage: `make shar`

### `packages`
- Builds all packages (Debian, RPM, SHAR).
- Depends on the `deb`, `rpm`, and `shar` targets.
- Usage: `make packages`

## Variables

- `BINARY_NAME`: The name of the output binary (default: `chperm`).
- `VERSION`: The version of the package to be built (default: `1.0.0`).
- `BUILD_ID`: The build ID, generated from the current Git commit hash.


## Continuous Integration

The `ci.yml` file is set up for CI/CD pipelines, triggering on tagged pushes and running on the latest Ubuntu version.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes.
4. Submit a pull request with a clear description of the changes.

## License

chperm is licensed under the MIT license