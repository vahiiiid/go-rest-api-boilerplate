#!/usr/bin/env python3
"""
Fix test assertions from direct response access to wrapped format
"""
import re

def main():
    filepath = "/Users/vahid.vahedi/grab/go-rest-api-boilerplate/internal/user/handler_test.go"
    
    with open(filepath, 'r') as f:
        content = f.read()
    
    # Fix pattern: assert.Equal(t, "message", response["message"]) in error contexts
    # Look for Status codes that indicate errors (400, 401, 403, 404, 409, 500)
    replacements = [
        # StatusConflict + "Email already exists"
        (
            r'(expectedStatus: http\.StatusConflict,\s+checkResponse: func\(t \*testing\.T, w \*httptest\.ResponseRecorder\) \{\s+var response map\[string\]interface\{\}\s+err := json\.Unmarshal\(w\.Body\.Bytes\(\), &response\)\s+assert\.NoError\(t, err\)\s+)assert\.Equal\(t, "Email already exists", response\["message"\]\)',
            r'\1assert.Equal(t, false, response["success"])\n\t\t\t\terrorInfo, ok := response["error"].(map[string]interface{})\n\t\t\t\tassert.True(t, ok, "error should be a map")\n\t\t\t\tassert.Equal(t, "Email already exists", errorInfo["message"])'
        ),
        # StatusInternalServerError + "database connection error"
        (
            r'(expectedStatus: http\.StatusInternalServerError,\s+checkResponse: func\(t \*testing\.T, w \*httptest\.ResponseRecorder\) \{\s+var response map\[string\]interface\{\}\s+err := json\.Unmarshal\(w\.Body\.Bytes\(\), &response\)\s+assert\.NoError\(t, err\)\s+)assert\.Equal\(t, "database connection error", response\["details"\]\)',
            r'\1assert.Equal(t, false, response["success"])\n\t\t\t\terrorInfo, ok := response["error"].(map[string]interface{})\n\t\t\t\tassert.True(t, ok, "error should be a map")\n\t\t\t\tassert.Equal(t, "database connection error", errorInfo["details"])'
        ),
    ]
    
    for pattern, replacement in replacements:
        content = re.sub(pattern, replacement, content, flags=re.DOTALL)
    
    # Generic pattern for all response["field"] in error contexts
    # This is tricky - let's just do string replacements for common patterns
    
    # For error responses: response["message"], response["details"], response["code"]
    # Need to wrap in response["error"]["field"]
    
    with open(filepath, 'w') as f:
        f.write(content)
    
    print(f"Fixed {filepath}")

if __name__ == "__main__":
    main()
