# SURF

Free Text Search across your infrastructure platforms via CLI.

S.U.R.F is an acronym for: `Search-Unified-Recursive-Fast` 

![image info](./docs/xs-logo.png)

# Supported Platforms

- [x] [Vault](https://www.vaultproject.io/)
- [X] [AWS Route53](https://github.com/Isan-Rivkin/route53-cli)
- [X] [AWS ACM](https://aws.amazon.com/certificate-manager/)
- [ ] Kubernetes - TODO  
- [ ] Consul - TODO

# AWS Route53 Usage 

Based on [AWS Route53](https://github.com/Isan-Rivkin/route53-cli): Search what's behind domain `api.my-corp.com`: 

```bash 
surf r53 -q api.my-corp.com
```

# AWS ACM Usage 

Search inside ACM Certificates in AWS.

Example search: containing a domain: 

```bash
surf acm -q my-domain.com
```

Example search: certificate attached to a loab balancer: 

```bash
surf acm -q 's:elasticloadbalancing:us-west-2:123:loadbalancer/app/alb' --filter-used-by
```

# Hashicorp Vault Usage 

Search the query `aws` in Vault: 

```bash
surf vault -q aws 
```

Configure a default mount to start search from in Vault: 

```bash
export SURF_VAULT_DEFAULT_MOUNT=<my-default-mount>
```

Store LDAP auth on your OS keychain: 

```bash
surf config
```

# Hashicorp Consul Usage

Search under the `scripts` path for keys ending with `.sh`

```bash
surf consul --prefix scripts --query "\.fish$" --output-url=false
```

# Install 

### Brew 

MacOS (and ubuntu supported) installation via Brew:

```bash
brew tap isan-rivkin/toolbox
brew install surf
```

### Download Binary

1. [from releases](https://github.com/Isan-Rivkin/surf/releases)

2. Move the binary to global dir and change name to `surf`:

```bash
cd <downloaded zip dir>
mv surf /usr/local/bin
```

### Install from Source

```bash
git clone git@github.com:Isan-Rivkin/surf.git
cd surf
go run main.go
```

# Supported Auth methods per platform

*Please open a PR and request additional methods if you need.*

## Authentication Methods 

- [x] Vault - LDAP 
- [x] AWS - via profile on `~/.aws/credentials file`
- [x] Consul - None


# Version check 

The CLI will query [github.com](https://github.com/Isan-Rivkin/surf/releases) to check if there is a newer version and print out a message to the terminal.

If you wish to opt out set the environment variable `SURF_VERSION_CHECK=false`. 

No Data is collected it is purely [github.com](https://github.com/Isan-Rivkin/surf/releases) query.



