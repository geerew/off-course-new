#!/bin/bash

inkscape=/Applications/Inkscape.app/Contents/MacOS/inkscape
outdir=output

while [[ $# -gt 0 ]]; do
  case $1 in
    -in)
      IN_PATH=$2
      shift 2
      ;;
    -ico)
      ICO=$2
      shift 2
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

mkdir $outdir

for sz in 16 24 32 48 128 256 512
do
    echo "[+] Generete ${sz}x${sz} png..."
    $inkscape --export-filename ${outdir}/icon_${sz}x${sz}.png -w $sz -h $sz $IN_PATH
    $inkscape --export-filename ${outdir}/icon_${sz}x${sz}@2x.png -w $((sz*2)) -h $((sz*2))$IN_PATH
done

if [ -n "$ICO" ]; then
    echo "[+] Generate .ico file..."
    magick png32:${outdir}/icon_16x16.png png32:${outdir}/icon_24x24.png png32:${outdir}/icon_32x32.png png32:${outdir}/icon_48x48.png png32:${outdir}/icon_128x128.png png32:${outdir}/icon_256x256.png png32:${outdir}/icon_512x512.png -colors 256 ${outdir}/${ICO}.ico
fi