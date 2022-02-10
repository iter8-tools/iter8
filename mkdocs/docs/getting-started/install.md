---
template: main.html
title: Install Iter8
hide:
- toc
---

# Install Iter8

=== "Brew"
    Install the latest stable release of the Iter8 CLI using `brew` as follows.

    ```shell
    brew tap iter8-tools/iter8
    brew install iter8
    ```
    
    ??? note "Install a specific version"
        You can install the latest stable release of the Iter8 CLI with specific major and minor version numbers. For example, the following command installs the latest stable release of the Iter8 CLI with major `0` and minor `8`.
        ```shell
        brew tap iter8-tools/iter8
        brew install iter8@0.8
        ```

=== "Binaries"
    Pre-compiled Iter8 binaries for many platforms are available [here](https://github.com/iter8-tools/iter8/releases). Uncompress the iter8-X-Y.tar.gz archive for your platform, and move the `iter8` binary to any folder in your PATH.

=== "Source"
    Build Iter8 from source as follows. Go `1.17+` is a pre-requisite.
    ```shell
    # you can replace master with a specific tag such as v0.8.29
    export REF=master
    https://github.com/iter8-tools/iter8.git?ref=$REF
    cd iter8
    make install
    ```

=== "Go 1.17+"
    Install the latest stable release of the Iter8 CLI using `go 1.17+` as follows.

    ```shell
    go install github.com/iter8-tools/iter8@latest
    ```
    You can now run `iter8` (from your gopath bin/ directory)

    ??? note "Install a specific version"
        You can also install Iter8 CLI with a specific tag. For example, the following command installs version `0.8.29` of the Iter8 CLI.
        ```shell
        go install github.com/iter8-tools/iter8@v0.8.29
        ```


