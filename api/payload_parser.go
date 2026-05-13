package main

import (
	"errors"

	"github.com/buger/jsonparser"
)

var errInvalidJSON = errors.New("invalid json")

func parsePayload(body []byte) (Payload, error) {
	var p Payload

	// Parse only the fields used by vectorization. Missing keys default to zero values.
	var err error
	if p.ID, err = getString(body, "id"); err != nil {
		return p, err
	}

	if p.Transaction.Amount, err = getFloat(body, "transaction", "amount"); err != nil {
		return p, err
	}
	var installments int64
	if installments, err = getInt(body, "transaction", "installments"); err != nil {
		return p, err
	}
	p.Transaction.Installments = int(installments)
	if p.Transaction.RequestedAt, err = getString(body, "transaction", "requested_at"); err != nil {
		return p, err
	}

	if p.Customer.AvgAmount, err = getFloat(body, "customer", "avg_amount"); err != nil {
		return p, err
	}
	var txCount int64
	if txCount, err = getInt(body, "customer", "tx_count_24h"); err != nil {
		return p, err
	}
	p.Customer.TxCount24h = int(txCount)
	if p.Customer.KnownMerchants, err = getStringArray(body, "customer", "known_merchants"); err != nil {
		return p, err
	}

	if p.Merchant.ID, err = getString(body, "merchant", "id"); err != nil {
		return p, err
	}
	if p.Merchant.MCC, err = getString(body, "merchant", "mcc"); err != nil {
		return p, err
	}
	if p.Merchant.AvgAmount, err = getFloat(body, "merchant", "avg_amount"); err != nil {
		return p, err
	}

	if p.Terminal.IsOnline, err = getBool(body, "terminal", "is_online"); err != nil {
		return p, err
	}
	if p.Terminal.CardPresent, err = getBool(body, "terminal", "card_present"); err != nil {
		return p, err
	}
	if p.Terminal.KmFromHome, err = getFloat(body, "terminal", "km_from_home"); err != nil {
		return p, err
	}

	lastRaw, lastType, _, err := jsonparser.Get(body, "last_transaction")
	if err != nil && !errors.Is(err, jsonparser.KeyPathNotFoundError) {
		return p, errInvalidJSON
	}
	if err == nil {
		// Accept object or null; anything else is invalid.
		switch lastType {
		case jsonparser.Object:
			var last LastTransaction
			if last.Timestamp, err = getString(lastRaw, "timestamp"); err != nil {
				return p, err
			}
			if last.KmFromCurrent, err = getFloat(lastRaw, "km_from_current"); err != nil {
				return p, err
			}
			p.LastTransaction = &last
		case jsonparser.Null:
			// leave nil
		default:
			return p, errInvalidJSON
		}
	}

	return p, nil
}

func getString(body []byte, keys ...string) (string, error) {
	value, err := jsonparser.GetString(body, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return "", nil
		}
		return "", errInvalidJSON
	}
	return value, nil
}

func getFloat(body []byte, keys ...string) (float64, error) {
	value, err := jsonparser.GetFloat(body, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, nil
		}
		return 0, errInvalidJSON
	}
	return value, nil
}

func getInt(body []byte, keys ...string) (int64, error) {
	value, err := jsonparser.GetInt(body, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, nil
		}
		return 0, errInvalidJSON
	}
	return value, nil
}

func getBool(body []byte, keys ...string) (bool, error) {
	value, err := jsonparser.GetBoolean(body, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return false, nil
		}
		return false, errInvalidJSON
	}
	return value, nil
}

func getStringArray(body []byte, keys ...string) ([]string, error) {
	value, dataType, _, err := jsonparser.Get(body, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return nil, nil
		}
		return nil, errInvalidJSON
	}
	if dataType != jsonparser.Array {
		return nil, errInvalidJSON
	}

	items := make([]string, 0, 4)
	var arrayErr error
	_, arrayErr = jsonparser.ArrayEach(value, func(item []byte, itemType jsonparser.ValueType, _ int, _ error) {
		if itemType != jsonparser.String {
			arrayErr = errInvalidJSON
			return
		}
		items = append(items, string(item))
	})
	if arrayErr != nil {
		return nil, errInvalidJSON
	}
	return items, nil
}
