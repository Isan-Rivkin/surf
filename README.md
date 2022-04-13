# SURF

Free Text Search across your infrastructure platforms via CLI.

S.U.R.F is an acronym for: `Search-Unified-Recursive-Fast` 

![image info](./docs/xs-logo.png)

# Supported Platforms

- [x] [Vault](https://www.vaultproject.io/)
- [ ] Kubernetes - WIP  
- [ ] AWS Route53 - WIP  
- [ ] Consul - WIP 

# Vault Usage 

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

# Supported Auth methods per platform

*Please open a PR and request additional methods if you need.*

## Vault

- [x] LDAP 
- [ ] Approle 
- [ ] AWS 
- [ ] Token 

# Version check 

The CLI will query [github.com](https://github.com/Isan-Rivkin/surf/releases) to check if there is a newer version and print out a message to the terminal.

If you wish to opt out set the environment variable `SURF_VERSION_CHECK=false`. 

No Data is collected it is purely [github.com](https://github.com/Isan-Rivkin/surf/releases) query.



