package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("TODO wshub")

	router := gin.Default()

	_ = router.Run(":8585")
}
