package schema

type AddExpenseRequest struct {
	Expense Expense `json:"expense"`
}

type RemoveExpenseRequest struct {
	ExpenseId ExpenseId `json:"expenseId"`
}

type GetExpensesRequest struct {
	Counterparty UserId `json:"counterparty"`
}

type GetExpenseRequest struct {
	Id ExpenseId `json:"id"`
}
