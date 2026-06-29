package org

// AreaMetricsResponse is the response for the area-metrics endpoint.
type AreaMetricsResponse struct {
	NodeID            string                `json:"nodeId"`
	EmployeeCount     int                   `json:"employeeCount"`
	EmployeesWithGoals int                  `json:"employeesWithGoals"`
	AvgProgress       *float64              `json:"avgProgress"`
	CompletedGoals    int                   `json:"completedGoals"`
	PendingGoals      int                   `json:"pendingGoals"`
	AvgRating         *float64              `json:"avgRating"`
	RatingsCount      int                   `json:"ratingsCount"`
	Employees         []AreaMetricsEmployee `json:"employees"`
}

// AreaMetricsEmployee is a light employee projection for area metrics.
type AreaMetricsEmployee struct {
	ID                 string `json:"id"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	ProfileID          string `json:"profileId"`
	ProfileName        string `json:"profileName"`
	ProfileDescription string `json:"profileDescription"`
	JobTitle           string `json:"jobTitle"`
}
