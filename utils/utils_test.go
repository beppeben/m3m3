package utils

import "testing"

func TestUtils(t *testing.T) {
   	//t.Log("Hash: " + ComputeMd5("testinghashfunction"))
	//t.Log("Rand: " + RandString(32))
	err := SendEmail("mymail@gmail.com", "test subject", "oh yeahhh")
	if err != nil {
		t.Log(err.Error())
	}
	
}