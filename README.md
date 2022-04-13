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


