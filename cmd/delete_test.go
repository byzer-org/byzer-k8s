package main

import (
	"fmt"
	"mlsql.tech/allwefantasy/deploy/pkg/utils"
	"testing"
)

func TestJson(t *testing.T) {
	query := utils.BuildJsonQueryFromStr(`
   {
  "items":[{"metadata":{"name":"jack"}}]
}
`)
	items, _ := query.Array("items")
	var podNames = make([]string, 0)
	for _, item := range items {
		v := item.(map[string]interface{})["metadata"].(map[string]interface{})["name"].(string)
		podNames = append(podNames, v)
	}

	for _, podName := range podNames {
		logger.Info(fmt.Sprintf("delete pod:%s", podName))
	}
}
