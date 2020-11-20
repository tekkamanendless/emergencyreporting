# New Authentication (December 6, 2020)

Required data:

```
CLIENT_ID=
CLIENT_SECRET=
USERNAME=
PASSWORD=
ER_AID=
ER_UID=
```

Testing:

```
TENANT_HOST=emergencyreportingb2crc.b2clogin.com
TENANT_SEGMENT=emergencyreportingb2crc.onmicrosoft.com
```

Production:

```
TENANT_HOST=login.emergencyreporting.com
TENANT_SEGMENT=login.emergencyreporting.com
```

```
curl "https://${TENANT_HOST}/${TENANT_SEGMENT}/B2C_1A_PasswordGrant/oauth2/v2.0/token" -X POST --data "grant_type=password&client_id=${CLIENT_ID}&client_secret=${CLIENT_SECRET}&scope=https://${TENANT_SEGMENT}/secure/full_access&response_type=token&username=${USERNAME}&password=${PASSWORD}&er_aid=${ER_AID}&er_uid=${ER_UID}"
```

