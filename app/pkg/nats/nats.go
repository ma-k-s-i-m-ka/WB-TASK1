package nats

import (
	"WBL0/app/internal/delivery"
	"WBL0/app/internal/item"
	"WBL0/app/internal/order"
	"WBL0/app/internal/payment"
	"WBL0/app/pkg/config"
	"WBL0/app/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"time"
)

func ConnectNATS(cfg config.Config) (*nats.Conn, error) {

	natsTimeout, natsCancel := context.WithTimeout(context.Background(), time.Duration(cfg.NATS.ConnectionTimeout)*time.Second)
	defer natsCancel()

	natsConn, err := nats.Connect(cfg.NATS.URL, nats.Timeout(time.Duration(cfg.NATS.ConnectionTimeout)*time.Second))
	if err != nil {
		return nil, fmt.Errorf("cannot parse NATS config from url %v", err)
	}

	if natsTimeout.Err() != nil {
		natsConn.Close()
		return nil, fmt.Errorf("connection to NATS canceled")
	}

	return natsConn, nil
}

func SubNATS(cfg config.Config, log logger.Logger, natsConn *nats.Conn, deliveryService delivery.Service, paymentService payment.Service, itemService item.Service, orderService order.Service) error {
	subject := cfg.NATS.SUB

	_, err := natsConn.Subscribe(subject, func(msg *nats.Msg) {
		data := msg.Data

		var dataMap map[string]json.RawMessage

		if err := json.Unmarshal(data, &dataMap); err != nil {
			fmt.Println("Ошибка разбора JSON:", err)
			return
		}

		var (
			createDeliveryDTO delivery.CreateDeliveryDTO
			createPaymentDTO  payment.CreatePaymentDTO
			createItemDTO     []*item.CreateItemDTO
			createOrderDTO    order.Order
		)

		if err := json.Unmarshal(dataMap["delivery"], &createDeliveryDTO); err != nil {
			fmt.Printf("Error decoding delivery JSON: %v\n", err)
			return
		}

		if err := json.Unmarshal(dataMap["payment"], &createPaymentDTO); err != nil {
			fmt.Printf("Error decoding payment JSON: %v\n", err)
			return
		}

		if err := json.Unmarshal(dataMap["items"], &createItemDTO); err != nil {
			fmt.Printf("Error decoding item JSON: %v\n", err)
			return
		}

		if err := json.Unmarshal(data, &createOrderDTO); err != nil {
			fmt.Printf("Error decoding order JSON: %v\n", err)
			return
		}

		delivery, err := deliveryService.Create(context.Background(), &createDeliveryDTO)
		if err != nil {
			fmt.Printf("Error creating delivery: %v\n", err)
			return
		}
		log.Info("Input Delivery ID: ", delivery.ID)

		payment, err := paymentService.Create(context.Background(), &createPaymentDTO)
		if err != nil {
			fmt.Printf("Error creating payment: %v\n", err)
			return
		}
		log.Info("Input Payment ID: ", payment.ID)

		itemIDs := []int64{}
		items, err := itemService.Create(context.Background(), createItemDTO)
		if err != nil {
			fmt.Printf("Error creating item: %v\n", err)
			return
		}
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
			log.Info("Input Items ID: ", item.ID)
		}

		crtord := order.CreateOrderDTO{
			OrderUID:          createOrderDTO.OrderUID,
			TrackNumber:       createOrderDTO.TrackNumber,
			Entry:             createOrderDTO.Entry,
			Delivery:          delivery.ID,
			Payment:           payment.ID,
			Items:             itemIDs,
			Locale:            createOrderDTO.Locale,
			InternalSignature: createOrderDTO.InternalSignature,
			CustomerID:        createOrderDTO.CustomerID,
			DeliveryService:   createOrderDTO.DeliveryService,
			ShardKey:          createOrderDTO.ShardKey,
			SMID:              createOrderDTO.SMID,
			DateCreated:       createOrderDTO.DateCreated,
			OofShard:          createOrderDTO.OofShard,
		}

		order, err := orderService.Create(context.Background(), &crtord)
		if err != nil {
			fmt.Printf("Error creating order: %v\n", err)
			return
		}
		log.Info("Input Order ID: ", order.OrderUID)

		fmt.Printf("\n\nReceived message: \n%s", string(data))
	})

	if err != nil {
		return fmt.Errorf("cannot subscribe to NATS %v", err)
	}
	return nil
}
