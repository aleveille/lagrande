package generator

func float64ToByteArrPtr(f float64) *[]byte {
	// Convert the float to an ascii representation []byte array
	// Instead of converting the float to a string and then the string to a byte array, we go through each digit and set the ascii value in the byte array
	// We have a fixed precision of 4 digits on generated floats for now, which makes this easier

	// Keep two local temp values, tempValueForDigitCount will be used to find how many digits there's to the left of the decimal point.
	tempValueForDigitCount := int64(f)
	// tempValue will be used to go through each digit
	tempValue := int64(f * 10000)

	byteArrayLen := 5 // Start with a byte array len of 5 which is the decimal point + 4 digits: .0000
	if f < 0 {        // If the value is negative, increase the byte array len by one to account for the minus sign we'll add
		byteArrayLen++
		tempValue = 0 - tempValue
	}
	for dowhile := true; dowhile; dowhile = tempValueForDigitCount != 0 { // Do while makes sure we at least do this once to account for numbers where (-1 < n < 1)
		tempValueForDigitCount /= 10
		byteArrayLen++
	}

	// Create the ascii byte array of length which we just calculated
	byteArr := make([]byte, byteArrayLen)

	// Set the four decimals
	byteArr[byteArrayLen-1] = byte(tempValue%10 + 48)
	tempValue /= 10
	byteArr[byteArrayLen-2] = byte(tempValue%10 + 48)
	tempValue /= 10
	byteArr[byteArrayLen-3] = byte(tempValue%10 + 48)
	tempValue /= 10
	byteArr[byteArrayLen-4] = byte(tempValue%10 + 48)
	tempValue /= 10

	byteArr[byteArrayLen-5] = 0x2E // This is ascii decimal point .
	// Set all digits left to the decimal point
	for i := 6; i <= byteArrayLen; i++ {
		byteArr[byteArrayLen-i] = byte(tempValue%10 + 48)
		tempValue /= 10
	}

	// Set the minus sign if appropriate
	if f < 0 {
		byteArr[0] = 0x2D
	}

	return &byteArr
}

func intToByteArrPtr(n int) *[]byte {
	// Convert the float to an ascii representation []byte array
	// Instead of converting the int to a string and then the string to a byte array, we go through each digit and set the ascii value in the byte array

	tempValue := n
	byteArrayLen := 0 // Start with a byte array len of 5 which is the decimal point + 4 digits: .0000
	if n < 0 {        // If the value is negative, increase the byte array len by one to account for the minus sign we'll add
		byteArrayLen++
	}
	for dowhile := true; dowhile; dowhile = tempValue != 0 { // Do while makes sure we at least do this once to account for numbers where (-1 < n < 1)
		tempValue /= 10
		byteArrayLen++
	}

	// Create the ascii byte array of length which we just calculated
	byteArr := make([]byte, byteArrayLen)

	if n >= 0 {
		tempValue = n // Reset temp value to n
	} else {
		tempValue = -n
	}
	// Set all digits left to the decimal point
	for i := 1; i <= byteArrayLen; i++ {
		byteArr[byteArrayLen-i] = byte(tempValue%10 + 48)
		tempValue /= 10
	}

	// Set the minus sign if appropriate
	if n < 0 {
		byteArr[0] = 0x2D
	}

	return &byteArr
}
