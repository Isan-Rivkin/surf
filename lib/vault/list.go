package vault

import (
	"fmt"
	"path/filepath"
	"strings"

	vaultApi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type NodeType string

const (
	Secret NodeType = "secret"
	Folder NodeType = "folder"
)

type Node struct {
	T           NodeType
	KeyValue    string
	BaseKeyPath string
}

func NewNode(n, base string) *Node {
	t := Secret
	if strings.HasSuffix(n, "/") {
		t = Folder
	}
	return &Node{
		T:           t,
		KeyValue:    n,
		BaseKeyPath: base,
	}
}

func (n *Node) GetFullPath() string {
	return filepath.Join(n.BaseKeyPath, n.KeyValue)
}

func (v *Vaultclient[A]) assemblePath(isList bool, p string) (string, error) {
	c, err := v.getClient()
	if err != nil {
		return "", err
	}
	p, err = AssemblePath(p, c)
	return p, err
}

// List keys including mounts, filter out non secrets / non expandable paths
// by expandable I mean things like secret engines, they are not expandable
func (v *Vaultclient[A]) ListTreeFiltered(basePath string) ([]*Node, error) {
	var nodes []*Node

	basePath, err := v.assemblePath(true, basePath)

	if err != nil {
		return nil, err
	}

	if IsRootPath(basePath) {
		mounts, err := v.ListMounts()

		if err != nil {
			return nil, fmt.Errorf("failed listing mounts for path %s in list filter %s", basePath, err.Error())
		}

		for m, v := range mounts {
			if IsStorage(v.Type) {
				nodes = append(nodes, NewNode(m, ""))
			}
		}
	} else {
		return v.ListTree(basePath)
	}
	return nodes, nil
}

func (v *Vaultclient[A]) ListMounts() (map[string]*vaultApi.MountOutput, error) {
	// get authenticated client
	client, err := v.getClient()

	if err != nil {
		return nil, err
	}
	mounts, err := client.Sys().ListMounts()
	return mounts, err
}

func (c *Vaultclient[A]) ListTree(basePath string) ([]*Node, error) {
	var nodes []*Node
	//get authenticated client
	client, err := c.getClient()

	if err != nil {
		return nil, err
	}

	keys, err := client.Logical().List(basePath)

	if err != nil {
		return nil, fmt.Errorf("failed listing base path %s", basePath)
	}

	if keys == nil {
		return nodes, nil
	}
	if _, found := keys.Data["keys"]; !found || keys == nil {
		e := fmt.Errorf("no keys under path %s", basePath)
		log.WithError(e).Error("failed in listing tree")
		return nil, e
	}

	folders, err := SecretToListOfStr(keys)

	if err != nil {
		return nil, err
	}

	for _, f := range folders {
		n := NewNode(f, basePath)
		nodes = append(nodes, n)
	}
	return nodes, nil
}
