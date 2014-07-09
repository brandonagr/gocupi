gocupi
======

Polargraph (vertical plotter / drawing machine) written in Go. Inspired by [Polargraph](http://www.polargraph.co.uk/) and [drawbot](http://marginallyclever.com/blog/drawbot/) projects, designed with the help of the [Dallas Makerspace](http://dallasmakerspace.org/)

The difference with Gocupi is that it uses a Raspberry Pi to do most of the processing instead of relying on a microcontroller to parse commands. This gives it the ability to render complex svg files that would not git in memory of a microcontroller.

Beta versions of prebuilt hardware are now available at [gocupi.com](http://gocupi.com)

Quickstart Guide
================

Install Go
-----------
The full install guide is on [golang.org](http://golang.org/doc/install)

For the raspberry pi you can [download precompiled binaries](http://dave.cheney.net/unofficial-arm-tarballs) to avoid the lengthy process of building go from source
  
Set your GOPATH environment variable
`cd ~`
`mkdir gopath`
`export GOPATH=$HOME/gopath`

Update your PATH environment variable
`export PATH=$PATH:$HOME/gopath/bin`

Download and build gocupi
---------------------------
`go get github.com/brandonagr/gocupi`

Run gocupi
----------
`gocupi` The gocupi executable should now be in your $GOPATH\bin folder, you can run the command with no arguments to display the help and what command are available

`gocupi setup 1000 700 700` The setup command can be used to initialize the dimensions of the polargraph hardware, the setup is stored in the generated gocupi_config.xml file

`gocupi -toimage grid 100 10` The -toimage flag causes the system to draw to an output.png instead of trying to control the stepper motors over serial

Basic polargraph description
============================
Two stepper motors move a pen hanging from threads to draw stuff out on a whiteboard or any vertical surface. A program written in Go runs on the Pi, it sends movement commands over serial to an arduino, which then pulses the step pin on the stepper drivers to make the stepper motors move.

This project is different from most other Polargraphs in that there is no step generation code on the arduino, everything is calculated in Go and then the arduino just receives a stream of step deltas that it stores in a memory buffer and then executes. Since all logic is written in Go running on the Pi it allows using more advanced interpolation models for smooth drawing, not needing to use fixed point or single precision floats for calculations, not needing to reflash the arduino often when making code changes, etc.

Design description
==================
In the Go program, there are several channels that form a pipeline where separate functions execute the different pipeline stages. All of the stages are run in different goroutines so that they execute concurrently.

* The first stage in the pipeline is generating X,Y coordinates, it can either read those points from an svg file, gcode file, mouse data, or generate them according to an algorithm(such as hilbert space filling curve, spiral, circle, parabolic graph, etc).

* The second stage takes an X,Y coordinate and interpolates the movement from the previous X,Y position to the new position by evaluating the pen position every 2 milliseconds. It takes into account acceleration, entry speed, and exit speed so that it can slow down the pen smoothly before the end of the current line segment if needed. It calculates how much the stepper motors need to turn to move the pen to the interpolated X,Y location over those 2 milliseconds.

* The final stage takes the step commands and writes them over serial to the arduino. The arduino first sends a byte requesting a certain amount of data when the buffer has enough room, then the raspberry pi sends that much data to the arduino.

* The arduino has a 1KB buffer of step commands and uses simple linear interpolation to see if it should generate a pulse at the current time to move the stepper motor one step in a particular direction.

In order to generate single line art drawings(like the raspberry pi logo shown below) I followed the [makerbot](http://www.makerbot.com/blog/2012/03/12/single-line-art-traveling-salesman-problem-tutorial/) and [eggbot](http://wiki.evilmadscience.com/TSP_art) tutorials which show how to convert a grayscale image to a stippled image to a path to an svg file.
