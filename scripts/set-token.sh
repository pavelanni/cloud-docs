#!/usr/bin/env bash
# This scripts uses the 1Password CLI to get the token and set it in the environment variable.

# Get the token from 1Password
TOKEN=$(op read op://Training/cloud-docs/token)

# Set the token in the environment variable
export TOKEN

# Print the token
echo "Token set in the environment variable: TOKEN"