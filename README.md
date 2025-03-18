
### Running Locally

#### Setup
Install the `entr` package, which enables the `make watch` command. This command
will watch for any changes to `*.go` files and build the Go installation,
much like `npm run dev` in Javascript.

#### Provider Installation
Setting up Terraform to use a local provider can be a tad tricky.
Copy the .env.dist file to .env and set the environment variables,
the relevant one being `TERRAFORM_PROVIDER_EXECUTABLE_LOCATION`.
Additionally, the TerraformRC must be set. On UNIX systems, this
file will be located at 
```hcl
provider_installation {
  dev_overrides {
   "registry.terraform.io/kassett/balena" = "~/.terraform.d/plugins/kassett/balena"
  }
}
```

#### Local Debugging
In Goland or IDEA, right-click on the play button next to `func main()`, click on Debug
and enter 
`-gcflags="all=-N -l" -o <PATH-TO-Plugin`' in `Go tool arguments`, as well as `--debug=True` in `Program arguments`.
The terminal will output an environment variable starting with `TF_REATTACH_PROVIDERS`.
When you go to run the terraform command again, preface with the same environment variable.