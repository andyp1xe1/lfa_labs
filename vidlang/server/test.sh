#!/bin/bash

videos=( "./screencast.mp4" "./screencast2.mp4" )

while true; do
  for video in "${videos[@]}"; do
    echo $video
    cp "$video" video.mp4
    sleep 3
  done
done
