package metadata

import (
	"fmt"

	"github.com/openshift/library-go/pkg/certs/cert-inspection/certgraphapi"
)

func newDescriptionViolation(name string, pkiInfo *certgraphapi.PKIRegistryInfo) (Violation, error) {
	registry := &certgraphapi.PKIRegistryInfo{}

	for i := range pkiInfo.CertKeyPairs {
		curr := pkiInfo.CertKeyPairs[i]
		description := curr.CertKeyInfo.Description
		if len(description) == 0 {
			registry.CertKeyPairs = append(registry.CertKeyPairs, curr)
		}
	}
	for i := range pkiInfo.CertificateAuthorityBundles {
		curr := pkiInfo.CertificateAuthorityBundles[i]
		description := curr.CABundleInfo.Description
		if len(description) == 0 {
			registry.CertificateAuthorityBundles = append(registry.CertificateAuthorityBundles, curr)
		}
	}

	v := Violation{
		Name:     name,
		Registry: registry,
	}

	markdown, err := generateMarkdownNoDescription(registry)
	if err != nil {
		return v, err
	}
	v.Markdown = markdown

	return v, nil
}

func generateMarkdownNoDescription(pkiInfo *certgraphapi.PKIRegistryInfo) ([]byte, error) {
	certsWithoutDescription := map[string]certgraphapi.PKIRegistryInClusterCertKeyPair{}
	caBundlesWithoutDescription := map[string]certgraphapi.PKIRegistryInClusterCABundle{}

	for i := range pkiInfo.CertKeyPairs {
		curr := pkiInfo.CertKeyPairs[i]
		owner := curr.CertKeyInfo.OwningJiraComponent
		description := curr.CertKeyInfo.Description
		if len(description) == 0 && len(owner) != 0 {
			certsWithoutDescription[owner] = curr
			continue
		}
	}
	for i := range pkiInfo.CertificateAuthorityBundles {
		curr := pkiInfo.CertificateAuthorityBundles[i]
		owner := curr.CABundleInfo.OwningJiraComponent
		description := curr.CABundleInfo.Description
		if len(description) == 0 && len(owner) != 0 {
			caBundlesWithoutDescription[owner] = curr
			continue
		}
	}

	md := NewMarkdown("Certificate Description")
	if len(certsWithoutDescription) > 0 || len(caBundlesWithoutDescription) > 0 {
		md.Title(2, fmt.Sprintf("Missing Description (%d)", len(certsWithoutDescription)+len(caBundlesWithoutDescription)))
		if len(certsWithoutDescription) > 0 {
			md.Title(3, fmt.Sprintf("Certificates (%d)", len(certsWithoutDescription)))
			md.OrderedListStart()
			for owner, curr := range certsWithoutDescription {
				md.NewOrderedListItem()
				md.Textf("ns/%v secret/%v\n", curr.SecretLocation.Namespace, curr.SecretLocation.Name)
				md.Textf("**JIRA component:** %v", owner)
				md.Text("\n")
			}
			md.OrderedListEnd()
			md.Text("\n")
		}
		if len(caBundlesWithoutDescription) > 0 {
			md.Title(3, fmt.Sprintf("Certificate Authority Bundles (%d)", len(caBundlesWithoutDescription)))
			md.OrderedListStart()
			for owner, curr := range caBundlesWithoutDescription {
				md.NewOrderedListItem()
				md.Textf("ns/%v configmap/%v\n", curr.ConfigMapLocation.Namespace, curr.ConfigMapLocation.Name)
				md.Textf("**JIRA component:** %v", owner)
				md.Text("\n")
			}
			md.OrderedListEnd()
			md.Text("\n")
		}
	}
	return md.Bytes(), nil
}

func diffCertKeyPairDescription(actual, expected certgraphapi.PKIRegistryCertKeyPairInfo) error {
	if actual.Description != expected.Description {
		return fmt.Errorf("expected description to be %s, but was %s", expected.Description, actual.Description)
	}
	return nil
}

func diffCABundleDescription(actual, expected certgraphapi.PKIRegistryCertificateAuthorityInfo) error {
	if actual.OwningJiraComponent != expected.OwningJiraComponent {
		return fmt.Errorf("expected description to be %s, but was %s", expected.Description, actual.Description)
	}
	return nil
}
