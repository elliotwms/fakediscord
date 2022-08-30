package snowflake

import "github.com/bwmarrin/snowflake"

var node *snowflake.Node

func Configure(nodeID int64) (err error) {
	node, err = snowflake.NewNode(nodeID)

	return err
}

func Generate() snowflake.ID {
	if node == nil {
		panic("snowflake.Generate called before snowflake.Configure")
	}

	return node.Generate()
}
