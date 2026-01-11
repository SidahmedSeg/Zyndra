-- OTP codes table for email verification
CREATE TABLE IF NOT EXISTS otp_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    purpose VARCHAR(50) NOT NULL DEFAULT 'registration', -- 'registration', 'password_reset', 'login'
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    
    -- Prevent duplicate active OTPs for same email/purpose
    CONSTRAINT unique_active_otp UNIQUE (email, purpose, code)
);

-- Index for quick lookup
CREATE INDEX IF NOT EXISTS idx_otp_email_purpose ON otp_codes(email, purpose);
CREATE INDEX IF NOT EXISTS idx_otp_expires ON otp_codes(expires_at);

-- Cleanup function to delete expired OTPs (can be called periodically)
-- In production, set up a cron job or scheduled task to run:
-- DELETE FROM otp_codes WHERE expires_at < now();

