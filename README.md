# stOrk Migration Tool

stOrk is a database migration tool.

## Installation

### Unix users (Linux, BSDs and MacOSX)

Unix users may download and install latest *stOrk* release with command:

```bash
sh -c "$(curl https://sweetohm.net/dist/stork/install)"
```

If *curl* is not installed on your system, you might use *wget* instead, running:

```bash
sh -c "$(wget -O - https://sweetohm.net/dist/stork/install)"
```

**Note:** Some directories are protected, even as *root*, on **MacOSX** (since *El Capitan* release), thus you can't install *stOrk* in */usr/bin* for instance.

### Binary package

Otherwise, you can download latest binary archive at <https://github.com/c4s4/stork/releases>. Unzip the archive, put the binary of your platform somewhere in your *PATH* and rename it *stork*.

### Gophers

If a recent version of Go is installed on your system, you may install *stOrk* with:

```
$ go get github.com/c4s4/stork
```

## Usage

To run migration scripts in *sql* directory, type:

```
$ stork sql
```

## Docker

You can pull docker image with:

```
$ docker pull casa/stork:latest
```

Then you can run it with:

```
$ docker run --network host --rm --volume=$(pwd)/sql:/sql --env-file=.env casa/stork /sql
```

Where:

- Migration scripts live in *sql* directory, so that:
  - We bind them in the */sql* directory of the container with `--volume=$(pwd)/sql:/sql`
  - We pass */sql* on command line to stork running in the container
- Database access are in *.env* file in current directory

The dotenv file to access database should look like:

```
MYSQL_HOSTNAME=localhost
MYSQL_USERNAME=stork
MYSQL_PASSWORD=stork
MYSQL_ROOT_PASSWORD=root
```

The user to run migration script must be granted rights to create database and access stork and targets databases.

Note that the docker image is **very** small (less than 5 MB, almost the same size than the binary) and performance running in docker is almost the same than running binary on command line.

*Enjoy!*
