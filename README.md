# stOrk Migration Tool

stOrk is a database migration tool.

## Installation

### Unix users (Linux, BSDs and MacOSX)

Unix users may download and install latest *stOrk* release with command:

```bash
sh -c "$(curl https://sweetohm.net/dist/stork/install)"
```

If *curl* is not installed on you system, you might run:

```bash
sh -c "$(wget -O - https://sweetohm.net/dist/stork/install)"
```

**Note:** Some directories are protected, even as *root*, on **MacOSX** (since *El Capitan* release), thus you can't install *stOrk* in */usr/bin* for instance.

### Binary package

Otherwise, you can download latest binary archive at <https://github.com/c4s4/stork/releases>. Unzip the archive, put the binary of your platform somewhere in your *PATH* and rename it *stork*.

## Usage

To run migration scripts in *sql* directory, type:

```
$ stork sql
```

*Enjoy!*
