# domainGPT

## Overview
`domainGPT` is a powerful domain name generator leveraging the capabilities of OpenAI's GPT-3 model. With `domainGPT`, you can create unique and catchy domain name suggestions based on either a specific name or a project idea. Furthermore, it verifies the availability of the suggested domains.

## Features
- Generates domain names based on a specific name or project idea.
- Verifies the availability of suggested domains.
- Easily extensible with a list of desired TLDs.

## Usage

### Setup
1. Clone the repository. 
2. Create a `.env` file in the root directory and add your OpenAI key: `OPENAI={YOUR_KEY}`. 
3. run `go get` 

### Commands
1. Generate domains based on an existing name: 
   `go run main.go name [YOUR_NAME]` 
2. Generate based off of project idea: 
   `go run main.go idea [YOUR_PROJECT_IDEA]` 

## Optional Flags
- `--skip-verify` or `-sv`: Use this flag if you want to skip the verification of domain names before returning them.
