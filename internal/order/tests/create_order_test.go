package tests

import (
	"context"
	"fmt"
	"log"
	"testing"

	sw "github.com/PIGcanstudy/gorder/common/client/order"
	_ "github.com/PIGcanstudy/gorder/common/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	ctx    = context.Background()
	server = fmt.Sprintf("http://%s/api", viper.GetString("order.http-addr"))
)

func TestMain(m *testing.M) {
	before() // 在所有测试运行之前进行一些初始化工作
	m.Run()  // 这个函数用于实际运行测试
}

func before() {
	log.Printf("server=%s", server)
}

func TestCreateOrder_success(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items: []sw.ItemWithQuantity{
			{
				Id:       "test-item-1",
				Quantity: 1,
			},
		},
	})
	t.Logf("body=%s", string(response.Body))
	assert.Equal(t, 200, response.StatusCode())

	assert.Equal(t, 0, response.JSON200.Errno)
}

func TestCreateOrder_invalidParams(t *testing.T) {
	response := getResponse(t, "123", sw.PostCustomerCustomerIdOrdersJSONRequestBody{
		CustomerId: "123",
		Items:      nil,
	})
	assert.Equal(t, 200, response.StatusCode())
	assert.Equal(t, 2, response.JSON200.Errno)
}

func getResponse(t *testing.T, customerId string, body sw.PostCustomerCustomerIdOrdersJSONRequestBody) *sw.PostCustomerCustomerIdOrdersResponse {
	t.Helper()
	client, err := sw.NewClientWithResponses(server)
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.PostCustomerCustomerIdOrdersWithResponse(ctx, customerId, body)
	if err != nil {
		t.Fatal(err)
	}
	return response
}
