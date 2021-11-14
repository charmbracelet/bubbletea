#!/bin/bash

for folder in /Users/Rod/Documents/github-repo/bubbletea/examples/*
  do
    singlename="${folder##*/}"
    echo -e "#$singlename Example\n![$singlename Recording](recording-$singlename.gif)" > "$folder/README.md"
    echo "$folder/README.md"
    echo $singlename
    echo -e "#$singlename Example\n![$singlename Recording](recording-$singlename.gif)"
  done
