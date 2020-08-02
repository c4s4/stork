# stOrk Migration Tool

*stOrk* is a database migration tool for MySQL.

## Installation

You can install binary for your platform with:

```bash
# using curl
sh -c "$(curl https://sweetohm.net/dist/stork/install)"
# using wget
sh -c "$(wget -O - https://sweetohm.net/dist/stork/install)"
```

Or you can pull docker image with:

```
$ docker pull casa/stork:latest
```

## Usage

To run migration scripts (starting with figures and ending with *.sql*) in current directory, type:

```
$ stork
```

To get help, type:

```
$ stork -help
Usage: stork [-env=file] [-init] [-dry] [-mute] [-white] [-version] dir
-env=file  Dotenv file to load
-init      Run all scripts
-dry       Dry run (won't execute scripts)
-mute      Don't print logs
-white     Don't print color
-version   Print version and exit
dir        Directory of migration scripts
```

You can specify directory where are migration scripts passing it on command line.

Access to MySQL database should be passed with environment variables. You may put them in a dotenv file, such as:

```
MYSQL_HOSTNAME=localhost
MYSQL_USERNAME=stork
MYSQL_PASSWORD=stork
```

You indicate this dotenv file with *-env=file* option.

The user to run migration script must be granted rights to create database and access stork and target databases.

With *-init* option, *stOrk* will erase table where are stored scripts that were already passed, and thus all migration scripts will run.

*-dry* will print scripts that should run but won't run them.

*-mute* will ask *stOrk* to print nothing on console except errors.

*-white* will disable color printing on console.

*-version* will print version on command line.

## Docker

To run *stOrk* with docker, you might type:

```
$ docker run --network host --rm --volume=$(pwd)/sql:/sql --env-file=.env casa/stork <options>
```

Where database access are in *.env* file in current directory.

Note that the docker image is **very** small (less than 5 MB, almost the size of the binary) and performance running in docker is almost the same than running binary on command line.

## Alternate installs

You may also download archive at <https://github.com/c4s4/stork/releases>. Unzip the archive, put the binary of your platform somewhere in your *PATH* and rename it *stork*.

If a recent version of Go is installed on your system, you may install *stOrk* with:

```
$ go get github.com/c4s4/stork
```

*Enjoy!*
