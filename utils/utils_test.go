package utils

import "testing"

func TestUtils(t *testing.T) {
   	//t.Log("Hash: " + ComputeMd5("testinghashfunction"))
	//t.Log("Rand: " + RandString(32))
	//err := SendEmail("mymail@gmail.com", "test subject", "oh yeahhh")
	//err := SaveImage("http://i.telegraph.co.uk/multimedia/archive/03454/mansell_v_3454166i.jpg", "test.jpg")
	
	err := DeleteAllImages()
	if err != nil {
		t.Log(err.Error())
	}
	
}