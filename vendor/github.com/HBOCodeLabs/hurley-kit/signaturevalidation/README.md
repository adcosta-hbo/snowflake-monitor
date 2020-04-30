# requestSignatureValidator

> A Golang version of [Hurley-Request-Signature-Validator](https://github.com/HBOCodeLabs/Hurley-Request-Signature-Validator) (Node version): generating request signature and validating signature functions

> NB: HMACs are an authenticity check, which is one function of digital signatures. 
> The other is attestation. Since signatures are created with a private key and validated with a shareable public key, 
> they attest to which party actually created the signature. HMACs, because they rely on a shared key between parties, can't provide this.
> Please refer to #security-team for any discussion on this.

### Description

This library provides two functions which are used to create and validate a request signature produced in a specific format

### Usage
There are two functions exposed, (a) for the middleware (b) for the generation of the request header.

### Generation and Validation

##### Generation

```
    signatureHeader, err := httputils.CreateSignedRequestHeader(c.purchaseSignature, requestBody)
    
    if err != nil {
    	rw.WriteHeader(http.StatusInternalServerError)
    	c.stats.Incr(metricName+errorMetricSuffix, 1)
    	tagSpanWithError(span, http.StatusInternalServerError, err)
    	ctxLog.Error("event", "createSignedRequestHeaderError", "error", err.Error())
    	return
    }
    		
    req.Header.Add(SignatureHeaderName, signatureHeader)
    resp, err := c.httpClient.Do(req)
    
    if err != nil {
        return nil, err
    }
    var responseBody interface{}
    
    if err := parseResponse(resp, &responseBody, true); err != nil {
        return nil, err
    }
    
    return &responseBody, nil
```
##### Validation

```
    signatureValidatorHandler, err := NewMiddleware(next, WithSecret("some-secret"))
    if err != nil {
      // do something
    }   
```  