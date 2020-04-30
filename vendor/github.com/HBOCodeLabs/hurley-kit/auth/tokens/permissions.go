package tokens

const (
	// The permissions are referred from the following location
	// its basically the copy of the permissions in the below link
	// https://github.com/HBOCodeLabs/Hurley-AuthCheck/blob/master/lib/Permissions.js
	// IMPORTANT: DO NOT CHANGE ANY OF THESE VALUES!

	// PermissionUpdateCreditCard End-User permission to update the credit card
	PermissionUpdateCreditCard = 1
	// PermissionChangePassword End-User Permission to change password
	PermissionChangePassword = 2
	// PermissionUpdateProfile End-User Permission to update profile
	PermissionUpdateProfile = 3
	// PermissionPlayVideo End-User Permission to play video
	PermissionPlayVideo = 4
	// PermissionReadCatalog is a General Permissions
	PermissionReadCatalog = 5
	// PermissionTimeTravel represents the permission enum that allows a user to
	// set a "server time" that is different from the current time.
	PermissionTimeTravel = 6
	// PermissionPlayFreeVideo General Permissions to play free video
	PermissionPlayFreeVideo = 7
	// PermissionManageDevice Device Permissions
	PermissionManageDevice = 8
	// PermissionRegisterAccount Registration permission
	PermissionRegisterAccount = 9
	// PermissionRegionFree as of Jan2018 - implemented for Mosaic but not longer used. Anticipate use for international.
	PermissionRegionFree = 10
	// PermissionAdminWriteEditorial Libby write permission
	PermissionAdminWriteEditorial = 11
	// PermissionAdminReadUserProfile Admin (Customer Service) Tier1 Permissions for csSearch.userSearch, accounts.getUserProfile
	PermissionAdminReadUserProfile = 12
	// PermissionAdminReadUserEntitlements Admin (Customer Service) Tier1 Permissions for billing.getEntitlementHistory, billing.receiptIdLookup
	PermissionAdminReadUserEntitlements = 13
	// PermissionAdminReadUserEvents Admin (Customer Service) Tier1 Permissions for greatlakes.getUserEvents
	PermissionAdminReadUserEvents = 14
	// PermissionAdminUpdateUserProfile Admin (Customer Service) Tier1 Permissions for accounts.updateUserProfile, accounts.passwordResetManual
	PermissionAdminUpdateUserProfile = 15
	// PermissionAdminAuthNRevocation Admin (Customer Service) Tier1 Permissions for tokenRevocations.authnRevocation (does a force logout on all devices)
	PermissionAdminAuthNRevocation = 16
	// PermissionAdminManageUserEntitlements Tier2 permission for billing.unlinkReceipt, billing.grantTemporaryEntitlement
	PermissionAdminManageUserEntitlements = 17
	// LAST USED PERMISSION NUMBER: 17
)
