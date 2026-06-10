package domains

import (
	"fmt"

	"github.com/spf13/cobra"
)

var domainsListCmd = &cobra.Command{Use: "list", Short: "List custom domains", RunE: runDomainsList}
var domainsCreateCmd = &cobra.Command{Use: "create", Short: "Create a custom domain", RunE: runDomainsCreate}
var domainsVerifyCmd = &cobra.Command{Use: "verify", Short: "Verify a custom domain", RunE: runDomainsVerify}
var domainsUpdateCmd = &cobra.Command{Use: "update", Short: "Update a custom domain hostname", RunE: runDomainsUpdate}
var domainsDeleteCmd = &cobra.Command{Use: "delete", Short: "Delete a custom domain", RunE: runDomainsDelete}

var (
	domainName    string
	domainID      string
	domainType    string
	domainNewName string
	domainConfirm bool
)

func init() {
	domainsCreateCmd.Flags().StringVar(&domainName, "name", "", "Hostname (required)")
	domainsCreateCmd.Flags().StringVar(&domainType, "type", "APPLICATION", "Application type: APPLICATION, FORMS, or FILES")
	domainsVerifyCmd.Flags().StringVar(&domainName, "name", "", "Hostname to verify (required)")
	domainsUpdateCmd.Flags().StringVar(&domainID, "domain", "", "Custom domain ID (required)")
	domainsUpdateCmd.Flags().StringVar(&domainNewName, "name", "", "New hostname (required)")
	domainsDeleteCmd.Flags().StringVar(&domainID, "domain", "", "Custom domain ID (required)")
	domainsDeleteCmd.Flags().BoolVarP(&domainConfirm, "confirm", "y", false, "Confirm deletion")
}

func runDomainsList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	companyID, err := client.ResolveCompanyID()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query CustomDomains($filter: CustomDomainFilterInput!, $skip: Int, $take: Int) { customDomains(filter: $filter, skip: $skip, take: $take) { items { %s } } }`, customDomainFields)
	var response struct {
		CustomDomains struct {
			Items []CustomDomain `json:"items"`
		} `json:"customDomains"`
	}
	variables := map[string]interface{}{"filter": map[string]interface{}{"companyId": companyID}, "skip": 0, "take": 50}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list domains: %w", err)
	}
	for _, d := range response.CustomDomains.Items {
		fmt.Printf("%s  %s  %s  %s\n", d.ID, d.ApplicationType, d.VerificationStatus, d.Name)
	}
	return nil
}

func runDomainsCreate(cmd *cobra.Command, args []string) error {
	if domainName == "" {
		return fmt.Errorf("domain name is required. Use --name")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation CreateCustomDomain($input: CreateCustomDomainInput!) { createCustomDomain(input: $input) { %s } }`, customDomainFields)
	variables := map[string]interface{}{"input": map[string]interface{}{"companyId": client.GetCompanyID(), "name": domainName, "applicationType": domainType}}
	var response struct {
		CreateCustomDomain CustomDomain `json:"createCustomDomain"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}
	return printJSON(response.CreateCustomDomain)
}

func runDomainsVerify(cmd *cobra.Command, args []string) error {
	if domainName == "" {
		return fmt.Errorf("domain name is required. Use --name")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	var response struct {
		VerifyCustomDomain bool `json:"verifyCustomDomain"`
	}
	if err := client.ExecuteQueryWithResult(`mutation VerifyCustomDomain($name: String!) { verifyCustomDomain(name: $name) }`, map[string]interface{}{"name": domainName}, &response); err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}
	fmt.Printf("Verified: %t\n", response.VerifyCustomDomain)
	return nil
}

func runDomainsUpdate(cmd *cobra.Command, args []string) error {
	if domainID == "" || domainNewName == "" {
		return fmt.Errorf("--domain and --name are required")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation UpdateCustomDomain($input: UpdateCustomDomainInput!) { updateCustomDomain(input: $input) { %s } }`, customDomainFields)
	variables := map[string]interface{}{"input": map[string]interface{}{"id": domainID, "name": domainNewName}}
	var response struct {
		UpdateCustomDomain CustomDomain `json:"updateCustomDomain"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	return printJSON(response.UpdateCustomDomain)
}

func runDomainsDelete(cmd *cobra.Command, args []string) error {
	if domainID == "" {
		return fmt.Errorf("domain ID is required. Use --domain")
	}
	if !domainConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	var response struct {
		DeleteCustomDomain bool `json:"deleteCustomDomain"`
	}
	if err := client.ExecuteQueryWithResult(`mutation DeleteCustomDomain($id: String!) { deleteCustomDomain(id: $id) }`, map[string]interface{}{"id": domainID}, &response); err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	fmt.Printf("Deleted: %t\n", response.DeleteCustomDomain)
	return nil
}
