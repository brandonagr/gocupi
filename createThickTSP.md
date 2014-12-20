Follow the instruction on http://www.drububu.com/illustration/tsp/index.html to create a tsp art svg image by using the following script. You have to have the correct folder structure if you use the script or you change the location of the tools in the script
acordingly.
The output saved in output_bilder\tsp_art.svg is the input for the gocupi meanderStipple command. For help see gocupi command help.

    # create a TSP art image according to the instructions on
    #http://www.drububu.com/illustration/tsp/index.html
    #Example
    #.\create_tsp.ps1 -stipp 1000 -pixd 6 -radm 2 -quali 0.1 -input .\input_bilder\test.png
    # .\create_tsp.ps1 -stipp 26000 -pixd 12 -radm 1 -quali 0.05 -input .\input_bilder\test.png
    # script created: 20.14.2014 / Sergio Daniels

    param(
      [int]$stipples = 8000,
      [string]$quality = 0.1,
      [string]$radmulti = 1,
      [string]$pixdensity = 5,
      [string]$inputFile = "input.png",
      [string]$output = ".\tmp\output.svg",
      [switch]$MyVerbose = $false
      )

    Write-Host "Stipples: ",$stipples
    Write-Host "color: ", $color
    Write-Host "Radius multiplikator: ", $radmulti
    Write-Host "qualitaet: ", $quality
    Write-Host "Subpixel densitiy: ", $pixdensity
    Write-Host "Eingabe-Datei: ", $inputFile
    Write-Host "Ausgabe-Datei: ", $output
    Write-Host "Verbose: ", $MyVerbose

    #create voronoi points (output.svg)

    # -f fixed radius
    .\voronoi\voronoi.exe -n -s $stipples -z $radmulti -p $pixdensity -t $quality $inputFile $output

    # if verbose
    if ($MyVerbose) {
      # do verbose stuff
      & 'C:\Program Files (x86)\Inkscape\inkscape.exe' $output
    }

    # create tsp format (positions.tsp)
    .\svg_extract $output
    mv -force positions.tsp .\tmp\positions.tsp

    # start Concorde select Heuritic -->Link kernigham-->when done exit concorde
    # save tour as .\tmp\tour.cyc
    & 'C:\Program Files (x86)\Concorde\Concorde.exe' .\tmp\positions.tsp| Out-Null

    #create tsp_art.svg
    .\tsp2svg $output .\tmp\tour.cyc
    mv -force tsp_art.svg .\output_bilder\tsp_art.svg

    #open inkscape
    #& 'C:\Program Files (x86)\Inkscape\inkscape.exe' .\tmp\tsp_art.svg
    # convert to png
    & 'C:\Program Files (x86)\Inkscape\inkscape.exe' -z  -h 1000 --file=.\output_bilder\tsp_art.svg -e .\output_bilder\tsp_art.png| Out-Null

    #open png file
    if ($MyVerbose) {
      & 'C:\Windows\System32\mspaint.exe' .\output_bilder\tsp_art.png
    }
