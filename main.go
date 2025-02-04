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
	totalDepositedAmount := 0
	for _, deposit := range deposits {
		totalDepositedAmount += deposit.Amount
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

	// calculate the total we need for onetime deposit
	oneTimeDepositExpectedAmount := 0
	for _, scheduledTransaction := range oneTimeDepositPlan.ScheduledTransactions {
		oneTimeDepositExpectedAmount += scheduledTransaction.Amount
	}

	// calculate the total we need for monthly deposit
	monthlyDepositExpectedAmount := 0
	for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
		monthlyDepositExpectedAmount += scheduledTransaction.Amount
	}

	// figure out how much we have to allocate for one time deposit, monthly deposit, and (if any) remaining
	onetimeDepositAmount, monthlyDepositAmount, remainder := getAmounts(
		totalDepositedAmount,
		oneTimeDepositExpectedAmount,
		monthlyDepositExpectedAmount,
	)

	if onetimeDepositAmount < oneTimeDepositExpectedAmount {
		// if insufficient, split proportionately across all portfolio
		remainingAmount := onetimeDepositAmount
		for _, scheduledTransaction := range oneTimeDepositPlan.ScheduledTransactions {
			proportion := float64(scheduledTransaction.Amount) / float64(oneTimeDepositExpectedAmount)
			proportionalAmount := int(math.Floor(float64(onetimeDepositAmount) * proportion))

			portfolio := portfolioByID[scheduledTransaction.PortfolioID]
			portfolio.Balance += proportionalAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount -= proportionalAmount
		}

		// remaining is possible as we're doing monetary division. just give it to the first portfolio as the amount is insignificant
		if remainingAmount > 0 {
			portfolioID := oneTimeDepositPlan.ScheduledTransactions[0].PortfolioID
			portfolio := portfolioByID[portfolioID]
			portfolio.Balance += remainingAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount = 0
		}
	} else {
		// else just split according to amount specified in transaction
		for _, scheduledTransaction := range oneTimeDepositPlan.ScheduledTransactions {
			portfolio := portfolioByID[scheduledTransaction.PortfolioID]
			portfolio.Balance += scheduledTransaction.Amount
			portfolioByID[portfolio.ID] = portfolio
		}
	}

	if monthlyDepositAmount < monthlyDepositExpectedAmount {
		// if insufficient, split proportionately across all portfolio
		remainingAmount := monthlyDepositAmount
		for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
			proportion := float64(scheduledTransaction.Amount) / float64(monthlyDepositExpectedAmount)
			proportionalAmount := int(math.Floor(float64(monthlyDepositAmount) * proportion))

			portfolio := portfolioByID[scheduledTransaction.PortfolioID]
			portfolio.Balance += proportionalAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount -= proportionalAmount
		}

		// remaining is possible as we're doing monetary division. just give it to the first portfolio as the amount is insignificant
		if remainingAmount > 0 {
			portfolioID := monthlyDepositPlan.ScheduledTransactions[0].PortfolioID
			portfolio := portfolioByID[portfolioID]
			portfolio.Balance += remainingAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount = 0
		}
	} else {
		// else just split according to amount specified in transaction
		for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
			portfolio := portfolioByID[scheduledTransaction.PortfolioID]
			portfolio.Balance += scheduledTransaction.Amount
			portfolioByID[portfolio.ID] = portfolio
		}
	}

	// split remainder amount proportionately across monthly deposits
	if remainder > 0 {
		remainingAmount := remainder
		for _, scheduledTransaction := range monthlyDepositPlan.ScheduledTransactions {
			proportion := float64(scheduledTransaction.Amount) / float64(monthlyDepositExpectedAmount)
			proportionalAmount := int(math.Floor(float64(remainder) * proportion))

			portfolio := portfolioByID[scheduledTransaction.PortfolioID]
			portfolio.Balance += proportionalAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount -= proportionalAmount
		}

		if remainingAmount > 0 {
			portfolioID := monthlyDepositPlan.ScheduledTransactions[0].PortfolioID
			portfolio := portfolioByID[portfolioID]
			portfolio.Balance += remainingAmount
			portfolioByID[portfolio.ID] = portfolio

			remainingAmount = 0
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

	return result
}

func getAmounts(totalDepositedAmount int, oneTimeDepositExpectedAmount int, monthlyDepositExpectedAmount int) (amountForOnetimeDeposit int, amountForMonthlyDeposit int, remainder int) {
	// if not deposit provided
	if totalDepositedAmount == 0 {
		return 0, 0, 0
	}

	// if not even sufficient for one time deposit plan
	if totalDepositedAmount < oneTimeDepositExpectedAmount {
		return totalDepositedAmount, 0, 0
	}

	totalExpectedAmount := oneTimeDepositExpectedAmount + monthlyDepositExpectedAmount
	// if not sufficient for all deposit plans
	if totalDepositedAmount < totalExpectedAmount {
		amountForMonthlyDepositPlan := totalDepositedAmount - oneTimeDepositExpectedAmount

		return totalDepositedAmount, amountForMonthlyDepositPlan, 0
	}

	// if there's extra
	if totalDepositedAmount > totalExpectedAmount {
		remainder = totalDepositedAmount - totalExpectedAmount

		return totalDepositedAmount, monthlyDepositExpectedAmount, remainder
	}

	// if deposit provided is the exact amount required by all deposit plans
	return oneTimeDepositExpectedAmount, monthlyDepositExpectedAmount, 0
}
