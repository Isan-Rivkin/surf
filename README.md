# SURF

Free Text Search across your infrastructure platforms via CLI.

It's like `... | grep` but for your entire infrastructure!

S.U.R.F is an acronym for: `Search-Unified-Recursive-Fast` 


![image info](./docs/xs-logo.png)

# Supported Platforms

- [x] [Hashicorp Vault](https://www.vaultproject.io/)
- [X] [AWS Route53](https://github.com/Isan-Rivkin/route53-cli)
- [X] [AWS ACM](https://aws.amazon.com/certificate-manager/)
- [X] [Hashicorp Consul KV](https://www.consul.io/docs/dynamic-app-config/kv)
- [ ] Kubernetes - TODO  

# Table of Contents 

- [Overview](#overview)
- [Usage Examples](#usage-examples)
  * [AWS Route53 Usage](#aws-route53-usage)
  * [AWS ACM Usage](#aws-acm-usage)
  * [Hashicorp Vault Usage](#hashicorp-vault-usage)
  * [Hashicorp Consul Usage](#hashicorp-consul-usage)
- [Install](#install)
    + [Brew](#brew)
    + [Download Binary](#download-binary)
    + [Install from Source](#install-from-source)
- [Authentication](#authentication)
  * [Supported Authentication Methods](#supported-authentication-methods)
- [Version check](#version-check)
- [How it Works](#how-it-works)

<small><i><a href='http://ecotrust-canada.github.io/markdown-toc/'>Table of contents generated with markdown-toc</a></i></small>

# Overview

SURF is built for Infrastructure Engineers as a CLI tool that enables searching any pattern across different platforms. 
Usually, the results are returned with a direct web URL. 

The search process depends on the context, for example: if you're searching in Vault it'll pattern match against keys. Instead, if you're searching in Route53 AWS a DNS address it'll return links to the targets behind it (e.g Load balancer). 


# Usage Examples 

## AWS Route53 Usage 

Based on [AWS Route53](https://github.com/Isan-Rivkin/route53-cli): Search what's behind domain `api.my-corp.com`: 

```bash 
surf r53 -q api.my-corp.com
```

## AWS ACM Usage 

Search inside ACM Certificates in AWS.

Example search: containing a domain: 

```bash
surf acm -q my-domain.com
```

Example search: certificate attached to a loab balancer: 

```bash
surf acm -q 's:elasticloadbalancing:us-west-2:123:loadbalancer/app/alb' --filter-used-by
```

## Hashicorp Vault Usage 

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

## Hashicorp Consul Usage

Search all keys containing the substring `server` 

```bash
surf consul -q "server"
```

Search under the `scripts` path for keys ending with `.sh`

```bash
surf consul --prefix scripts --query "\.sh$"
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

# Authentication

*Please open a PR and request additional methods if you need.*

## Supported Authentication Methods 

- [x] Vault - LDAP (run `$surf config` )
- [x] AWS - via profile on `~/.aws/credentials file`
- [x] Consul - None


# Version check 

The CLI will query [github.com](https://github.com/Isan-Rivkin/surf/releases) to check if there is a newer version and print out a message to the terminal.

If you wish to opt out set the environment variable `SURF_VERSION_CHECK=false`. 

No Data is collected it is purely [github.com](https://github.com/Isan-Rivkin/surf/releases) query.


# How it Works 

![image info](./docs/surf-flow.png)
