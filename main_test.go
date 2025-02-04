package main

import (
	"reflect"
	"testing"
)

func Test_GetPortfolioFinalAmount(t *testing.T) {
	type Input struct {
		portfolios   []Portfolio
		depositPlans []DepositPlan
		deposits     []Deposit
	}
	type TestCase struct {
		name   string
		input  Input
		expect []Portfolio
	}

	// for convenience as it'll be re-used across test cases and remain unchanged
	var (
		portfolios = []Portfolio{
			{
				ID:   1,
				Name: "High risk",
			},
			{
				ID:   2,
				Name: "Retirement",
			},
		}

		depositPlans = []DepositPlan{
			{
				Type: DepositPlanType_OneTime,
				ScheduledTransactions: []ScheduledTransaction{
					{
						PortfolioID: 1,
						Amount:      10000,
					},
					{
						PortfolioID: 2,
						Amount:      500,
					},
				},
			},
			{
				Type: DepositPlanType_Monthly,
				ScheduledTransactions: []ScheduledTransaction{
					{
						PortfolioID: 1,
						Amount:      0,
					},
					{
						PortfolioID: 2,
						Amount:      100,
					},
				},
			},
		}
	)

	testCases := []TestCase{
		{
			name: "Original from assignment",
			input: Input{
				portfolios:   portfolios,
				depositPlans: depositPlans,
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        10500,
					},
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        100,
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 10000,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 600,
				},
			},
		},
		{
			name: "Insufficient amount deposited, should prioritise one-time deposits",
			input: Input{
				portfolios:   portfolios,
				depositPlans: depositPlans,
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        10500,
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 10000,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 500,
				},
			},
		},
		{
			name: "Insufficient amount deposited, should prioritise one-time deposits. If any remain, credit to monthly deposits",
			input: Input{
				portfolios:   portfolios,
				depositPlans: depositPlans,
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        10501,
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 10000,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 501,
				},
			},
		},
		{
			name: "Insufficient amount deposited (for even a single scheduled transaction). Just deposit whatever we have proportionately to one time deposits",
			input: Input{
				portfolios: portfolios,
				depositPlans: []DepositPlan{
					{
						Type: DepositPlanType_OneTime,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      4000,
							},
							{
								PortfolioID: 2,
								Amount:      1000,
							},
						},
					},
					{
						Type: DepositPlanType_Monthly,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      0,
							},
							{
								PortfolioID: 2,
								Amount:      100,
							},
						},
					},
				},
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        100,
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 80,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 20,
				},
			},
		},
		{
			name: "More than required amount deposited. Remaining should be credited proportionately",
			input: Input{
				portfolios: portfolios,
				depositPlans: []DepositPlan{
					{
						Type: DepositPlanType_OneTime,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      10000,
							},
							{
								PortfolioID: 2,
								Amount:      500,
							},
						},
					},
					{
						Type: DepositPlanType_Monthly,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      400,
							},
							{
								PortfolioID: 2,
								Amount:      100,
							},
						},
					},
				},
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        11000, // this simulates the first deposit that'll cover everything
					},
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        500, // this simulates the extra amount that should be distributed proportionately
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 10800,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 700,
				},
			},
		},
		{
			name: "More than required amount deposited. If there's remaining even after credited proportionately, credit to first porfolio",
			input: Input{
				portfolios: portfolios,
				depositPlans: []DepositPlan{
					{
						Type: DepositPlanType_OneTime,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      10000,
							},
							{
								PortfolioID: 2,
								Amount:      500,
							},
						},
					},
					{
						Type: DepositPlanType_Monthly,
						ScheduledTransactions: []ScheduledTransaction{
							{
								PortfolioID: 1,
								Amount:      400,
							},
							{
								PortfolioID: 2,
								Amount:      100,
							},
						},
					},
				},
				deposits: []Deposit{
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        11000, // this simulates the first deposit that'll cover everything
					},
					{
						ReferenceCode: "YN5XWAAQ",
						Amount:        501, // this simulates the extra amount that should be distributed proportionately + remainder
					},
				},
			},
			expect: []Portfolio{
				{
					ID:      1,
					Name:    "High risk",
					Balance: 10801,
				},
				{
					ID:      2,
					Name:    "Retirement",
					Balance: 700,
				},
			},
		},
	}

	for _, testCase := range testCases {
		result := GetPortfolioFinalAmount(testCase.input.portfolios, testCase.input.depositPlans, testCase.input.deposits)

		if !reflect.DeepEqual(result, testCase.expect) {
			t.Errorf("Test fail\nName:\t\t%s.\nExpect:\t\t%+v \nObserved:\t%+v\n", testCase.name, testCase.expect, result)
		}
	}
}
