package executor

//func TestAccountTree(t *testing.T) {
//	dir, statedb, localdb := util.CreateTestDB()
//	defer util.CloseTestDB(dir, statedb)
//	NewAccountTree(localdb)
//	tree, err := getAccountTree(localdb)
//	assert.Equal(t, nil, err)
//	assert.NotEqual(t, nil, tree)
//	for i := 0; i < 3000; i++ {
//		ethAddress := "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" + strconv.Itoa(i)
//		_, err := AddNewLeaf(localdb, ethAddress, "ETH", 1, 1000)
//		assert.Equal(t, nil, err)
//	}
//	tree, err = getAccountTree(localdb)
//	t.Log(tree.RootMap)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int32(3000), tree.GetTotalIndex())
//	for i := 0; i < 10; i++ {
//		_, err = UpdateLeaf(localdb, int32(i+1), "ETH", 1, 1000)
//		assert.Equal(t, nil, err)
//		tree, err = getAccountTree(localdb)
//		t.Log(tree.RootMap)
//		assert.Equal(t, nil, err)
//	}
//}
