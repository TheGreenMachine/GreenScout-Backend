# GreenScout Backend
The backend/server software for FRC team 1816's scouting app, GreenScout.

**IMPORTANT: IT'S FINE THAT SERVER.CRT AND SERVER.KEY ARE PUBLIC. THEY ARE ONLY USED FOR MOCKING HTTPS ON LOCALHOST FOR TESTING** (this will be fixed...)

## Getting Started

It's expected that you at least know basic programming principles and practices. Additionally, knowing the [Go Programming Language](https://go.dev/learn/) will be neccessary when contributing to this project. A (relatively) short primer on Go is provided [here](docs/Go.md).

[VS Code](https://code.visualstudio.com/Download) is the recommended editor of choice for this project, along with installing it's [Golang Extension](https://marketplace.visualstudio.com/items?itemName=golang.Go).

Regardless of your development enviroment, you will need:
- [Go](https://go.dev/dl/)
- [Git](https://git-scm.com/downloads)
- [Python](https://www.python.org/downloads/)-(Need Python3 on macOS) and [pip](https://pypi.org/project/pip/)-(Need pip3 on macOS)
- [Sqlite 3 (optional)](https://sqlite.org/download.html)-(Pre-downloaded on macOS)
- C/C++ toolchain for `cgo` bindings

> ⚠️ **C/C++ on Windows**
>
> Windows does not ship with native capabilities to compile C/C++ programs. That plus the absence of a built-in package manager means there's a bit of a process to get this working.
>
> We recommend following [this guide on Using GCC with MinGW](https://code.visualstudio.com/docs/cpp/config-mingw) to install and configure `mingw-w64` to set up an appropriate toolchain. Avoid `Cygwin`.

> 💡 **IDE of Choice**
>
> While the rest of this document will detail how to set up your environment from a command line terminal, feel free to continue by your own method of choice.

Once you have all that installed, you'll need to actually download the code. From the terminal, clone and open this git repo:

```bash
$ git clone https://github.com/TheGreenMachine/GreenScout-Backend.git
$ cd GreenScout-Backend
```

### Dependency management

This repo makes use of a `vendor/` directory for its Go dependencies. This ensures that anyone will be able to build and run the program even if some dependency (or the developer) falls offline. During development, if you need to update your dependencies, go to the terminal and do the following from the base of the repo:

```bash
$ go mod tidy && go mod vendor
```

### Backend initialization

To run through the setup:

```bash
$ go run main.go setup
```

For a thoughrough breakdown of the setup process, go [here](./docs/Setup.md)

Once you've finished that, you can run the server in production.

If you're on a mac or linux machine, ports 80 and 443 are bound to the root only, so you have to run this with `sudo`:

```bash
## Production
$ sudo go run main.go prod
## Testing
$ sudo go run main.go test
```

If you want it to write the match numbers of the configured event to the spreadsheet first

```bash
$ sudo go run main.go prod matches
```

### Important setup information
  - You will need to know how to port forward in order to ping the server from external networks.
  - You will need a valid domain name, as I could not find a way to get ACME autocert to work without it.

### More docs! MORE DOCS!

Additional documentation on various topics can be found [in the `docs` directory](./docs/). Technical documentation can be found in the exhaustively-annotated functions in this project, however.

## Roadmap (Things for future developers to add)
  * Admin Features
    * Removing assigned matches from users. (Mutable schedules - talk to Lydia for specifics of what she wants)
  * Discrepencies for multi-scouting - the only one that is implemented right now is average times being too different
  * Greenlogger improvement
    * Having the errors also spit out the line of code/method/stacktrace they came from

## Helm Repo Usage

[Helm](https://helm.sh) must be installed to use the charts.  Please refer to
Helm's [documentation](https://helm.sh/docs) to get started.

Once Helm has been set up correctly, add the repo as follows:

  helm repo add greenscout https://thegreenmachine.github.io/GreenScout-Backend

If you had already added this repo earlier, run `helm repo update` to retrieve
the latest versions of the packages.  You can then run `helm search repo
<alias>` to see the charts.

To install the greenscout chart:

  helm install my-<chart-name> greenscout/greenscout-backend

To uninstall the chart:

    helm uninstall my-<chart-name>