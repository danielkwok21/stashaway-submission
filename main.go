package main

import (
	"fmt"
)

// DepositPlan represents either onetime / monthly deposits to various portfolios
type DepositPlan struct {
	Type                  DepositPlanType
	ScheduledTransactions []ScheduledTransaction
}

// DepositPlanType is an enum to represent either onetime / monthly deposit plans
type DepositPlanType string

const (
	DepositPlanType_OneTime DepositPlanType = "onetime"
	DepositPlanType_Monthly DepositPlanType = "monthly"
)

// ScheduledTransaction specifies a portfolio's expected deposit amount
type ScheduledTransaction struct {
	// Favouring using name instead of ID for the sake of simplicity here. IRL this will be a table, and all relations will be managed with IDs
	PortfolioName string
	Amount        int
}

type Deposit struct {
	ReferenceCode string
	Amount        int
}

type Portfolio struct {
	PortfolioName string
	Balance       int
}

func main() {
	depositPlans := []DepositPlan{
		{
			Type: DepositPlanType_OneTime,
			ScheduledTransactions: []ScheduledTransaction{
				{
					PortfolioName: "High risk",
					Amount:        10000,
				},
				{
					PortfolioName: "Retirement",
					Amount:        500,
				},
			},
		},
		{
			Type: DepositPlanType_Monthly,
			ScheduledTransactions: []ScheduledTransaction{
				{
					PortfolioName: "High risk",
					Amount:        10,
				},
				{
					PortfolioName: "Retirement",
					Amount:        100,
				},
			},
		},
	}

	deposits := []Deposit{
		{
			ReferenceCode: "YN5XWAAQ",
			Amount:        10500,
		},
		{
			ReferenceCode: "YN5XWAAQ",
			Amount:        100,
		},
		{
			ReferenceCode: "YN5XWAAQ",
			Amount:        100,
		},
	}

	result := getPortfolioFinalAmount(depositPlans, deposits)
	fmt.Println(result)
}

func getPortfolioFinalAmount(depositPlans []DepositPlan, deposits []Deposit) []Portfolio {
	portfolioByName := map[string]Portfolio{}
	for _, depositPlan := range depositPlans {
		for _, tx := range depositPlan.ScheduledTransactions {
			portfolioByName[tx.PortfolioName] = Portfolio{
				PortfolioName: tx.PortfolioName,
				Balance:       0,
			}
		}
	}

	var oneTimeDepositPlan DepositPlan
	var monthlyDepositPlan DepositPlan
	for _, depositPlan := range depositPlans {
		switch depositPlan.Type {
		case DepositPlanType_OneTime:
			oneTimeDepositPlan = depositPlan
		case DepositPlanType_Monthly:
			monthlyDepositPlan = depositPlan
		default:
			continue
		}
	}

	totalAmountReceived := 0
	for _, deposit := range deposits {
		totalAmountReceived += deposit.Amount
	}
	for _, scheduledTransaction := range oneTimeDepositPlan.ScheduledTransactions {
		if totalAmountReceived <= 0 {
			break
		}
		portfolio := portfolioByName[scheduledTransaction.PortfolioName]

		if totalAmountReceived < scheduledTransaction.Amount {
			portfolio.Balance += totalAmountReceived
			totalAmountReceived = 0
		} else {
			portfolio.Balance += scheduledTransaction.Amount
			totalAmountReceived -= scheduledTransaction.Amount
		}

		portfolioByName[scheduledTransaction.PortfolioName] = portfolio
	}
	for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
		if totalAmountReceived <= 0 {
			break
		}
		portfolio := portfolioByName[scheduledTransaction.PortfolioName]

		if totalAmountReceived < scheduledTransaction.Amount {
			portfolio.Balance += totalAmountReceived
			totalAmountReceived = 0
		} else {
			portfolio.Balance += scheduledTransaction.Amount
			totalAmountReceived -= scheduledTransaction.Amount
		}

		portfolioByName[scheduledTransaction.PortfolioName] = portfolio
	}

	// handle remaining, distribute proportionally
	if totalAmountReceived > 0 {
		totalScheduledTransactionAmount := 0
		scheduledTransactionByPortfolioName := map[string]ScheduledTransaction{}
		for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
			totalScheduledTransactionAmount += scheduledTransaction.Amount

			scheduledTransactionByPortfolioName[scheduledTransaction.PortfolioName] = scheduledTransaction
		}

		for _, portfolio := range portfolioByName {
			scheduledTransaction := scheduledTransactionByPortfolioName[portfolio.PortfolioName]
			proportionalAmount := totalAmountReceived * scheduledTransaction.Amount / totalScheduledTransactionAmount

			portfolio.Balance += proportionalAmount
			portfolioByName[portfolio.PortfolioName] = portfolio
			totalAmountReceived -= proportionalAmount
		}
	}

	var portfolios []Portfolio
	for _, portfolio := range portfolioByName {
		portfolios = append(portfolios, portfolio)
	}

	// if there's still amount left, usually due to division math remainder
	if totalAmountReceived > 0 {
		portfolios[0].Balance += totalAmountReceived
	}

	return portfolios
}
