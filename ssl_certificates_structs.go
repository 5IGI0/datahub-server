package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

type SSLCertificateRow struct {
	Id          int64  `db:"id"`
	HashId      string `db:"hash_id"`
	Certificate []byte `db:"certificate"`
	RowVer      uint8  `db:"row_ver"`

	IssuerRFC4514 string         `db:"issuer_rfc4514"`
	IssuerName    sql.NullString `db:"issuer_name"`
	IssuerOrga    sql.NullString `db:"issuer_orga"`

	SubjectRFC4514 string         `db:"subject_rfc4514"`
	SubjectName    sql.NullString `db:"subject_name"`
	SubjectOrga    sql.NullString `db:"subject_orga"`

	ValidBefore string `db:"valid_before"`
	ValidAfter  string `db:"valid_after"`
	PublicKey   string `db:"public_key"`

	_DNSNames []string
}

func (r SSLCertificateRow) GetId() int64 { return r.Id }
func (r SSLCertificateRow) GetHashId() string {
	if r.HashId == "" {
		return r.CompHashId()
	}
	return r.HashId
}
func (r SSLCertificateRow) CompHashId() string {
	h := sha1.Sum(r.Certificate)
	return hex.EncodeToString(h[:])
}

func (r *SSLCertificateRow) FromMap(input map[string]any) error {
	{
		b64_cert, _ := input["raw"].(string)
		if b64_cert == "" {
			return errors.New("no `raw` key in certificate")
		}
		var err error

		r.Certificate, err = base64.StdEncoding.DecodeString(b64_cert)
		if err != nil {
			return errors.Join(errors.New("unable to parse `raw` in certificate"), err)
		}
		r.HashId = r.CompHashId()
	}

	// TODO: parse certificates server-side (so rows are consistent and we can easily add new fields)

	issuer := input["issuer"].(map[string]any)
	issuer_attrs, _ := issuer["attrs"].(map[string]any)
	subject := input["subject"].(map[string]any)
	subject_attrs, _ := subject["attrs"].(map[string]any)

	r.RowVer = 0
	r.IssuerRFC4514, _ = issuer["rfc4514"].(string)
	r.SubjectRFC4514, _ = subject["rfc4514"].(string)
	r.ValidBefore, _ = input["valid_before"].(string)
	r.ValidAfter, _ = input["valid_after"].(string)
	r.PublicKey, _ = input["public_key"].(string)

	dns_names := input["dns_names"].([]any)
	r._DNSNames = make([]string, 0, len(dns_names))

	for _, n := range dns_names {
		name, _ := n.(string)
		r._DNSNames = append(r._DNSNames, name)
	}

	if l, e := issuer_attrs["organizationName"].([]any); e && len(l) >= 1 {
		r.IssuerOrga.String, _ = l[0].(string)
		r.IssuerOrga.Valid = true
	}

	if l, e := issuer_attrs["commonName"].([]any); e && len(l) >= 1 {
		r.IssuerName.String, _ = l[0].(string)
		r.IssuerName.Valid = true
	}

	if l, e := subject_attrs["organizationName"].([]any); e && len(l) >= 1 {
		r.SubjectOrga.String, _ = l[0].(string)
		r.SubjectOrga.Valid = true
	}

	if l, e := subject_attrs["commonName"].([]any); e && len(l) >= 1 {
		r.SubjectName.String, _ = l[0].(string)
		r.SubjectName.Valid = true
	}

	r.SubjectRFC4514 = TruncateText(r.SubjectRFC4514, 255)
	r.SubjectName.String = TruncateText(r.SubjectName.String, 127)
	r.SubjectOrga.String = TruncateText(r.SubjectOrga.String, 127)

	r.IssuerRFC4514 = TruncateText(r.IssuerRFC4514, 255)
	r.IssuerName.String = TruncateText(r.IssuerName.String, 127)
	r.IssuerOrga.String = TruncateText(r.IssuerOrga.String, 127)

	return nil
}
