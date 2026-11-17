package routes

import (
	"master-service/models"
	"master-service/pkg/datasource"
)

// applyDataSourceSSL resolves ssl_mode (from conn.SSLMode, or the legacy conn.SSL flag when
// empty), keeps SSL in sync with it and validates the CA certificate.
//
// incomingRootCert is the certificate sent by the client. When it is empty the stored
// certificate on conn is kept for verify modes (the settings UI does not resend it on every
// save) and dropped for require/disable, so a stale CA never lingers. A certificate sent
// explicitly with require/disable is rejected rather than silently ignored.
func applyDataSourceSSL(conn *models.DataSourceConnection, incomingRootCert string) error {
	mode, err := datasource.NormalizeSSLMode(conn.SSL, conn.SSLMode)
	if err != nil {
		return err
	}
	if incomingRootCert != "" {
		conn.SSLRootCert = incomingRootCert
	} else if mode != datasource.SSLModeVerifyCA && mode != datasource.SSLModeVerifyFull {
		conn.SSLRootCert = ""
	}
	if err := datasource.ValidateSSLConfig(mode, conn.SSLRootCert); err != nil {
		return err
	}
	conn.SSLMode = mode
	conn.SSL = mode != datasource.SSLModeDisable
	return nil
}
