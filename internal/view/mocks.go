package view

var sg1Mock = map[string]any{
	"[blue]id[white]": "sg-123456",
	"[blue]resource[white]": map[string]any{
		"[blue]GroupDescription[white]": "K8S Clusters",
		"[blue]GroupId[white]":          "sg-123456",
		"[blue]GroupName[white]":        "my-server-asikojdasiodja",
		"[blue]Id[white]":               "sg-123456",
		"[blue]SecurityGroupEgress[white]": []any{
			map[string]any{
				"[blue]CidrIp[white]":      "0.0.0.0/0",
				"[blue]Description[white]": "Node all egress",
				"[blue]FromPort[white]":    -1,
				"[blue]IpProtocol[white]":  "-1",
				"[blue]ToPort[white]":      -1,
			},
			map[string]any{
				"[blue]CidrIpv6[white]":    "::/0",
				"[blue]Description[white]": "Node all egress",
				"[blue]FromPort[white]":    -1,
				"[blue]IpProtocol[white]":  "-1",
				"[blue]ToPort[white]":      -1,
			},
		},
		"[blue]SecurityGroupIngress[white]": []any{
			map[string]any{
				"[blue]Description[white]":                "Cluster API to node 6443/tcp webhook",
				"[blue]FromPort[white]":                   6443,
				"[blue]IpProtocol[white]":                 "tcp",
				"[blue]SourceSecurityGroupId[white]":      "sg-6544455666",
				"[blue]SourceSecurityGroupOwnerId[white]": "333444555",
				"[blue]ToPort[white]":                     6443,
			},
			map[string]any{
				"[blue]CidrIp[white]":      "20.50.196.0/23",
				"[blue]Description[white]": "kubernetes.io/rule/nlb/health=aaaa555okjindsfoisdjfi",
				"[blue]FromPort[white]":    31026,
				"[blue]IpProtocol[white]":  "tcp",
				"[blue]ToPort[white]":      31026,
			},
			map[string]any{
				"[blue]CidrIp[white]":      "20.50.194.0/23",
				"[blue]Description[white]": "kubernetes.io/rule/nlb/health=aaaa555okjindsfoisdjfi",
				"[blue]FromPort[white]":    31026,
				"[blue]IpProtocol[white]":  "tcp",
				"[blue]ToPort[white]":      31026,
			},
			map[string]any{
				"[blue]CidrIp[white]":      "20.50.192.0/23",
				"[blue]Description[white]": "kubernetes.io/rule/nlb/health=aaaa555okjindsfoisdjfi",
				"[blue]FromPort[white]":    31026,
				"[blue]IpProtocol[white]":  "tcp",
				"[blue]ToPort[white]":      31026,
			},
			map[string]any{
				"[blue]CidrIp[white]":      "0.0.0.0/0",
				"[blue]Description[white]": "kubernetes.io/rule/nlb/client=aaaa555okjindsfoisdjfi",
				"[blue]FromPort[white]":    31026,
				"[blue]IpProtocol[white]":  "tcp",
				"[blue]ToPort[white]":      31026,
			},
			// ... (continues with all remaining ingress rules in similar map structures)
		},
		"[blue]Tags[white]": []any{
			map[string]any{
				"[blue]Key[white]":   "Project",
				"[blue]Value[white]": "my-app-server",
			},
			map[string]any{
				"[blue]Key[white]":   "kubernetes.io/cluster/my-app-server",
				"[blue]Value[white]": "owned",
			},
			map[string]any{
				"[blue]Key[white]":   "Environment",
				"[blue]Value[white]": "staging",
			},
			map[string]any{
				"[blue]Key[white]":   "Name",
				"[blue]Value[white]": "my-app-server-node",
			},
		},
		"[blue]VpcId[white]": "vpc-123213123",
	},
	"[blue]type[white]": "AWS::EC2::SecurityGroup",
}
