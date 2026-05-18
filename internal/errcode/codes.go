package errcode

const (
	AuthMissingAuthorizationHeader = 40101
	AuthInvalidAuthorizationHeader = 40102
	AuthInvalidToken               = 40103
	AuthMissingUserContext         = 40104
	AuthInvalidRequest             = 40001
	AuthInvalidCredential          = 40110
	AuthUserDisabled               = 40310
	AuthLoginFailed                = 50010
	AuthUnauthorized               = 40111
	AuthUnauthorizedContext        = 40112
	AuthUserNotFound               = 40113
	AuthQueryUserFailed            = 50011
	AuthPermissionDenied           = 40301
)

const (
	AdminUserQueryFailed         = 50020
	AdminUserCreateInvalidInput  = 40020
	AdminUserCreateFailed        = 40021
	AdminUserUpdateInvalidInput  = 40022
	AdminUserNotFoundForUpdate   = 40420
	AdminUserUpdateFailed        = 40023
	AdminUserNotFoundForDisable  = 40421
	AdminUserDisableFailed       = 50021
	AdminUserResetInvalidInput   = 40024
	AdminUserNotFoundForReset    = 40422
	AdminUserResetPasswordFailed = 50022
)

const (
	AdminHostQueryFailed        = 50030
	AdminHostCreateInvalidInput = 40030
	AdminHostCreateFailed       = 40031
	AdminHostUpdateInvalidInput = 40032
	AdminHostNotFoundForUpdate  = 40430
	AdminHostUpdateFailed       = 40033
	AdminHostNotFoundForDelete  = 40431
	AdminHostDeleteFailed       = 50031
	AdminHostNotFoundForSync    = 40432
	AdminHostSyncFailed         = 50230
)

const (
	AdminVMQueryFailed               = 50040
	AdminVMCreateInvalidInput        = 40040
	AdminVMCreateUnauthorized        = 40120
	AdminVMCreateFailed              = 40041
	AdminVMDeleteUnauthorized        = 40121
	AdminVMNotFoundForDelete         = 40440
	AdminVMDeleteFailed              = 40042
	AdminVMUpdateConfigInvalidInput  = 40043
	AdminVMUpdateConfigUnauthorized  = 40122
	AdminVMNotFoundForUpdateConfig   = 40441
	AdminVMUpdateConfigFailed        = 40044
	AdminVMResizeInvalidInput        = 40045
	AdminVMResizeUnauthorized        = 40123
	AdminVMNotFoundForResize         = 40442
	AdminVMResizeFailed              = 40046
	AdminVMUpdateNetworkInvalidInput = 40047
	AdminVMUpdateNetworkUnauthorized = 40124
	AdminVMNotFoundForUpdateNetwork  = 40443
	AdminVMUpdateNetworkFailed       = 40048
	AdminVMAssignInvalidInput        = 40049
	AdminVMAssignUnauthorized        = 40125
	AdminVMNotFoundForAssign         = 40444
	AdminVMAssignFailed              = 40050
	AdminVMMigrateInvalidInput       = 40051
	AdminVMMigrateUnauthorized       = 40126
	AdminVMNotFoundForMigrate        = 40445
	AdminVMMigrateFailed             = 40052
	AdminVMActionUnauthorized        = 40127
	AdminVMNotFoundForAction         = 40446
	AdminVMActionFailed              = 40053
	AdminVMListImagesFailed          = 50240
)

const (
	UserVMListUnauthorized           = 40130
	UserVMListFailed                 = 50050
	UserVMDetailUnauthorized         = 40131
	UserVMNotFoundForDetail          = 40450
	UserVMDetailFailed               = 50051
	UserVMActionUnauthorized         = 40132
	UserVMNotFoundForAction          = 40451
	UserVMActionFailed               = 40060
	UserVMResourceUnauthorized       = 40133
	UserVMNotFoundForResource        = 40452
	UserVMResourceFailed             = 50250
	UserVMResetPasswordUnauthorized  = 40134
	UserVMNotFoundForResetPassword   = 40453
	UserVMResetPasswordFailed        = 50251
	UserVMUpdatePasswordInvalidInput = 40062
	UserVMUpdatePasswordFailed       = 50252
)

const (
	AdminSystemDashboardFailed  = 50060
	AdminSystemLogsFailed       = 50061
	AdminSystemVMNotFound       = 40460
	AdminSystemVMResourceFailed = 50260
)

const (
	VNCIssueTokenUnauthorized   = 40135
	VNCIssueTokenNotFound       = 40454
	VNCIssueTokenFailed         = 40061
	VNCMissingToken             = 40136
	VNCInvalidToken             = 40137
	VNCConnectBackendFailed     = 50270
	ShellIssueTokenUnauthorized = 40138
	ShellIssueTokenNotFound     = 40455
	ShellIssueTokenFailed       = 40063
	ShellMissingToken           = 40139
	ShellInvalidToken           = 40140
	ShellConnectBackendFailed   = 50271
)

const (
	RateLimitExceeded = 42901
)
