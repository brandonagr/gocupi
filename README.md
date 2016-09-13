Create Thick TSP
================
forked the code in order to draw thicker lines to get more contrast. For detailed descriptions see file createThickTSP


gocupi
======

Polargraph (vertical plotter / drawing machine) written in Go. Inspired by [Polargraph](http://www.polargraph.co.uk/) and [drawbot](http://marginallyclever.com/blog/drawbot/) projects, designed with the help of the [Dallas Makerspace](http://dallasmakerspace.org/)

Gocupi is different from existing systems in that it uses a Raspberry Pi to do most of the processing instead of relying on a microcontroller to parse commands. This gives it the ability to render complex svg files and patterns that would not fit in memory on a microcontroller.

**A kit with all the hardware needed is now [available at gocupi.com](http://www.gocupi.com/store/)**

Check out the [project page](http://brandonagr.github.io/gocupi/) for a general description, there are also several [wiki pages](https://github.com/brandonagr/gocupi/wiki) with additional information as well as a [forum](http://4um.gocupi.com).

Custom Linux Distro
===================

You can now [download](https://www.dropbox.com/s/6l89m04kcqiwlgb/gocupi_boot_image.zip?dl=0) a custom linux distrobution built from Raspbian with all gocupi software pre-installed.
Unzip the download and copy files to blank SD card (do not build the image). Put the card in the Pi and Bob's your uncle.

If you choose the method above (custom linux distro) then you can skip to the bottom of this document to section "Run gocupi".

Manual Installation Quickstart Guide
================

(Start with something well supported i.e. Raspbian Wheezy.) 

Install Go
-----------
The full install guide is on [golang.org](http://golang.org/doc/install). For the Raspberry Pi you can [download precompiled binaries](http://dave.cheney.net/unofficial-arm-tarballs) to avoid the lengthy process of building go from source.

For example, installing go to the default location with a precompiled binary on a raspberry pi just takes 3 commands:

    wget http://dave.cheney.net/paste/go.1.3.linux-arm~multiarch-armv6-1.tar.gz
    sudo tar -C /usr/local -xzf go.1.3.linux-arm~multiarch-armv6-1.tar.gz
    sudo echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile

(You must set GOROOT also if you install go to someplace besides the default /usr/local/go)

Fix the serial communication (for Raspberry Pi)
----------------------------
Change the default serial communication by editing /boot/cmdline.txt and remove references to `/dev/ttyAMA0`

Disable the getty on that serial port in /etc/inittab by commenting out references to `/dev/ttyAMA0`

In order to reset the serial communication port we need to `sudo reboot`

Setup gopath folder
----------------------------
Setup the folder for your gopath and set some path variables

    mkdir ~/gopath
    sudo echo 'export GOPATH=$HOME/gopath' >> ~/.profile
    sudo echo 'export PATH=$PATH:$HOME/gopath/bin' >> ~/.profile

Install needed dependency
---------------------
Update apt-get, install mercurial, and reboot

    sudo apt-get update
    sudo apt-get install mercurial
    sudo reboot

Download and build gocupi
---------------------------
From your home directory

    go get github.com/brandonagr/gocupi

Run gocupi
----------
`gocupi` The gocupi executable should now be in your $GOPATH\bin folder, you can run the command with no arguments to display the help and what command are available

`gocupi setup 1000 700 700` The setup command can be used to initialize the dimensions of the polargraph hardware, the setup is stored in the generated gocupi_config.xml file

`gocupi -toimage grid 100 10` The -toimage flag causes the system to draw to an output.png instead of trying to control the stepper motors over serial
