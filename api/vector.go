package main

import "time"

const (
	mccRiskDefault = 0.5
)

func GenerateVector(p Payload) ([14]float64, error) {
	var v [14]float64

	txTime, err := parseTimeUTC(p.Transaction.RequestedAt)
	if err != nil {
		return v, err
	}

	v[0] = clamp01(p.Transaction.Amount / NormalizationData.MaxAmount)
	v[1] = clamp01(float64(p.Transaction.Installments) / NormalizationData.MaxInstallments)

	avgAmount := p.Customer.AvgAmount
	if avgAmount <= 0 {
		v[2] = 1.0
	} else {
		ratio := (p.Transaction.Amount / avgAmount) / NormalizationData.AmountVsAvgRatio
		v[2] = clamp01(ratio)
	}

	v[3] = float64(txTime.Hour()) / 23.0
	v[4] = float64(normalizedWeekday(txTime)) / 6.0

	if p.LastTransaction == nil {
		v[5] = -1
		v[6] = -1
	} else {
		lastTime, err := parseTimeUTC(p.LastTransaction.Timestamp)
		if err != nil {
			return v, err
		}
		minutes := txTime.Sub(lastTime).Minutes()
		if minutes < 0 {
			minutes = 0
		}
		v[5] = clamp01(minutes / NormalizationData.MaxMinutes)
		v[6] = clamp01(p.LastTransaction.KmFromCurrent / NormalizationData.MaxKm)
	}

	v[7] = clamp01(p.Terminal.KmFromHome / NormalizationData.MaxKm)
	v[8] = clamp01(float64(p.Customer.TxCount24h) / NormalizationData.MaxTxCount24h)

	if p.Terminal.IsOnline {
		v[9] = 1
	}
	if p.Terminal.CardPresent {
		v[10] = 1
	}
	if !knownMerchant(p.Merchant.ID, p.Customer.KnownMerchants) {
		v[11] = 1
	}

	v[12] = mccRisk(p.Merchant.MCC)
	v[13] = clamp01(p.Merchant.AvgAmount / NormalizationData.MaxMerchantAvgAmount)

	return v, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func normalizedWeekday(t time.Time) int {
	// Go: Sunday=0. Desired: Monday=0, Sunday=6.
	return (int(t.Weekday()) + 6) % 7
}

func knownMerchant(id string, known []string) bool {
	for i := 0; i < len(known); i++ {
		if known[i] == id {
			return true
		}
	}
	return false
}

func mccRisk(mcc string) float64 {
	if MccRisk == nil {
		return mccRiskDefault
	}
	if risk, ok := MccRisk[mcc]; ok {
		return risk
	}
	return mccRiskDefault
}

func parseTimeUTC(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}
