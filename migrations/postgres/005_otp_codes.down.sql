-- Remove OTP codes table
DROP INDEX IF EXISTS idx_otp_expires;
DROP INDEX IF EXISTS idx_otp_email_purpose;
DROP TABLE IF EXISTS otp_codes;

