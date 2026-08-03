package common

func GetTrustQuota() int {
	// model-gateway policy: finite consumer tokens must reserve quota before
	// relay. The upstream 5,000,000-unit wallet trust bypass turns a token cap
	// into a post-settlement soft limit and can admit a request without a token
	// reservation. Returning zero makes BillingSession.shouldTrust false while
	// retaining the normal pre-consume insufficient-token-quota denial path.
	//
	// REQUIRES_REVIEW_BEFORE_PRODUCTION_USE: this affects every wallet-funded
	// token in this new-api process.
	return 0
}
