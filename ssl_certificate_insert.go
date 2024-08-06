package main

import "github.com/Masterminds/squirrel"

func SSLCertificateInsert(r SSLCertificateRow) int64 {
	update_id := int64(0)
	insert_id := InsertHashIdBasedRows([]SSLCertificateRow{r}, "ssl_certificates", nil,
		func(row SSLCertificateRow) map[string]any {
			return map[string]any{
				"certificate":     row.Certificate,
				"row_ver":         row.RowVer,
				"issuer_rfc4514":  row.IssuerRFC4514,
				"issuer_name":     row.IssuerName,
				"issuer_orga":     row.IssuerOrga,
				"subject_rfc4514": row.SubjectRFC4514,
				"subject_name":    row.SubjectName,
				"subject_orga":    row.SubjectOrga,
				"valid_before":    row.ValidBefore,
				"valid_after":     row.ValidAfter,
				"public_key":      row.PublicKey}
		}, func(_ SSLCertificateRow, id int64) map[string]interface{} {
			update_id = id
			return nil
		})

	if insert_id != 0 {
		for _, name := range r._DNSNames {
			q, v := squirrel.
				Insert("ssl_certificate_dns_names").
				SetMap(map[string]any{
					"certificate_id": insert_id,
					"domain":         name}).MustSql()
			GlobalContext.Database.MustExec(q, v...)
		}
		return insert_id
	}

	return update_id
}
