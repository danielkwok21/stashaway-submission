package main

import (
	"math"
	"sort"
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
	PortfolioID int
	Amount      int
}

// Deposit is the deposit we receive from bank
type Deposit struct {
	ReferenceCode string
	Amount        int
}

type Portfolio struct {
	ID      int
	Name    string
	Balance int
}

func GetPortfolioFinalAmount(portfolios []Portfolio, depositPlans []DepositPlan, deposits []Deposit) []Portfolio {
	// organise portfolio by id for easy read/write
	portfolioByID := map[int]Portfolio{}
	for _, portfolio := range portfolios {
		portfolioByID[portfolio.ID] = portfolio
	}

	// calculate the total amount of deposit we've received
	totalAmountReceived := 0
	for _, deposit := range deposits {
		totalAmountReceived += deposit.Amount
	}

	// separate out one time vs monthly deposits as we'll process one time deposits first
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

	// currentAmount will set to total amount first, and slowly be deducted as we allocate to different funds
	currentAmount := totalAmountReceived

	// first we handle one time deposits
	for _, scheduledTransaction := range oneTimeDepositPlan.ScheduledTransactions {
		if currentAmount <= 0 {
			break
		}

		// find the portfolio this scheduled transaction is targeting
		portfolio := portfolioByID[scheduledTransaction.PortfolioID]

		// in case the amount we have left is less than the amount we need, credit what ever is left to this portfolio
		if currentAmount < scheduledTransaction.Amount {
			portfolio.Balance += currentAmount
			currentAmount = 0
		} else {
			// else update balance according to scheduled transactions
			portfolio.Balance += scheduledTransaction.Amount
			currentAmount -= scheduledTransaction.Amount
		}

		// persist updated portfolio
		portfolioByID[scheduledTransaction.PortfolioID] = portfolio
	}

	// then we handle monthly deposits
	for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
		if currentAmount <= 0 {
			break
		}

		// find the portfolio this scheduled transaction is targeting
		portfolio := portfolioByID[scheduledTransaction.PortfolioID]

		// in case the amount we have left is less than the amount we need, credit what ever is left to this portfolio
		if currentAmount < scheduledTransaction.Amount {
			portfolio.Balance += totalAmountReceived
			currentAmount = 0
		} else {
			// else update balance according to scheduled transactions
			portfolio.Balance += scheduledTransaction.Amount
			currentAmount -= scheduledTransaction.Amount
		}

		// persist updated portfolio
		portfolioByID[scheduledTransaction.PortfolioID] = portfolio
	}

	// if there's money left,  distribute proportionally
	if currentAmount > 0 {
		leftoverAmount := currentAmount
		// used to calculate proportion later
		totalScheduledTransactionAmount := 0
		// create a map for easy read/write
		scheduledTransactionByPortfolioID := map[int]ScheduledTransaction{}
		for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
			totalScheduledTransactionAmount += scheduledTransaction.Amount

			scheduledTransactionByPortfolioID[scheduledTransaction.PortfolioID] = scheduledTransaction
		}

		for _, portfolio := range portfolioByID {
			// figure out what's the proportionate amount this portfolio should receive
			scheduledTransaction := scheduledTransactionByPortfolioID[portfolio.ID]
			proportion := float64(scheduledTransaction.Amount) / float64(totalScheduledTransactionAmount)
			proportionalAmount := int(math.Floor(float64(leftoverAmount) * proportion))

			portfolio.Balance += proportionalAmount
			currentAmount -= proportionalAmount

			portfolioByID[portfolio.ID] = portfolio
		}
	}

	// turn out map into array to return as result
	var result []Portfolio
	for _, portfolio := range portfolioByID {
		result = append(result, portfolio)
	}

	// sort by ID desc for consistent testing
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	// if there's still amount left, usually due to division math remainder. just create to the first portfolio
	if currentAmount > 0 {
		result[0].Balance += currentAmount
		currentAmount -= currentAmount
	}

	return result
}
