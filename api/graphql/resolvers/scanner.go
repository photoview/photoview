package resolvers

import (
	"context"

	"github.com/viktorstrate/photoview/api/graphql/models"
)

func (r *mutationResolver) ScanAll(ctx context.Context) (*models.ScannerResult, error) {
	// if err := scanner.ScanAll(r.Database); err != nil {
	// 	errorMessage := fmt.Sprintf("Error starting scanner: %s", err.Error())
	// 	return &models.ScannerResult{
	// 		Finished: false,
	// 		Success:  false,
	// 		Message:  &errorMessage,
	// 	}, nil
	// }

	// startMessage := "Scanner started"

	// return &models.ScannerResult{
	// 	Finished: false,
	// 	Success:  true,
	// 	Message:  &startMessage,
	// }, nil
	panic("not implemented")
}

func (r *mutationResolver) ScanUser(ctx context.Context, userID int) (*models.ScannerResult, error) {
	// if err := scanner.ScanUser(r.Database, userID); err != nil {
	// 	errorMessage := fmt.Sprintf("Error scanning user: %s", err.Error())
	// 	return &models.ScannerResult{
	// 		Finished: false,
	// 		Success:  false,
	// 		Message:  &errorMessage,
	// 	}, nil
	// }

	// startMessage := "Scanner started"

	// return &models.ScannerResult{
	// 	Finished: false,
	// 	Success:  true,
	// 	Message:  &startMessage,
	// }, nil
	panic("not implemented")
}
