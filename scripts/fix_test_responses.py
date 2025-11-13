#!/usr/bin/env python3
"""
Script to update test assertions from old response format to new wrapped format.
Old: {"id": 1, "name": "John"} 
New: {"success": true, "data": {"id": 1, "name": "John"}}
"""

import re
import sys

def fix_test_file(content):
    """Fix test assertions to use new response format"""
    
    # Pattern 1: assert.Contains(t, "VALUE", response["field"]) -> check in error
    # This is for error responses checking error.code
    content = re.sub(
        r'assert\.Contains\(t, "(\w+)", response\["code"\]\)',
        r'''assert.Equal(t, false, response["success"])
				errorInfo, ok := response["error"].(map[string]interface{})
				assert.True(t, ok, "error should be a map")
				assert.Equal(t, "\1", errorInfo["code"])''',
        content
    )
    
    # Pattern 2: assert.Equal(t, "value", response["message"]) -> check in error
    # For error responses
    def replace_error_message(match):
        indent = match.group(1)
        value = match.group(2)
        field = match.group(3)
        
        # Check context to see if this is error response
        return f'''{indent}assert.Equal(t, false, response["success"])
{indent}errorInfo, ok := response["error"].(map[string]interface{})
{indent}assert.True(t, ok, "error should be a map")
{indent}assert.Equal(t, "{value}", errorInfo["{field}"])'''
    
    # Pattern 3: Direct field access for success responses
    # assert.Equal(t, value, response["field"]) where field is id, name, email, etc.
    
    return content

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 fix_test_responses.py <test_file>")
        sys.exit(1)
    
    filename = sys.argv[1]
    
    with open(filename, 'r') as f:
        content = f.read()
    
    fixed_content = fix_test_file(content)
    
    with open(filename, 'w') as f:
        f.write(fixed_content)
    
    print(f"Fixed {filename}")

if __name__ == "__main__":
    main()
