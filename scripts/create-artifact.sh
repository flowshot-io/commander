#!/bin/bash

# Check if input and output arguments were provided
if [ $# -lt 2 ] || [ $# -gt 3 ]; then
  echo "Usage: $0 <input_dir> <output_file> [working_dir]"
  exit 1
fi

# Set the input directory and output file name
input_dir="$1"
output_file="$2"
working_dir="${3:-.}"

# Create the output file's parent directory if it doesn't exist
output_dir="$(dirname "$output_file")"
if [ ! -d "$output_dir" ]; then
  mkdir -p "$output_dir"
fi

# Tar and gzip the directory's files
tar -czvf "$output_file" -C "$working_dir" "$input_dir"

echo "Archive created: $output_file"