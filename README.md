# 42ActivityAPI

This repository contains APIs written in Go language that are provided to a system aimed at streamlining 42Tokyo's operations and visualizing student contributions.

## Overview
- Endpoints involved in school cleaning shift management.
- Endpoints involved in managing users.
- Endpoints involved in recording student contributions.
- Endpoints involved in managing NFC readers.

## Getting started
This API requires Go, docker, and the 42API UID and secret.

1. Clone the repository.
   ```
   git clone https://github.com/42association/42ActivityAPI.git
   cd 42ActivityAPI
   ```

2. Write environment variables to `.env`.
3. Execute the API.
   ```
   make up
   ```

## Usage

After executing the API, access the /apidoc endpoint.

## Contribution

Pull requests are always welcome, but if you're thinking of making major changes, please open an issue first to discuss it.