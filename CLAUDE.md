# CLAUDE.md

## Project: LogGuardian Go Lambda Function

You are helping build a Go-based AWS Lambda function that fixes non-compliant CloudWatch log groups.

## Working Guidelines

### **Research First, Code Second**
- Always research and use the **latest stable versions** of Go, AWS SDK v2, and libraries
- Look up current best practices before implementing
- Don't assume - verify syntax, function signatures, and patterns online

### **File Management**
- You have full permission to create, update, and delete files
- This is a git-based system - everything is versioned
- Proceed without asking for permission to make changes
- Create proper Go module structure (`go.mod`, proper directory layout)

### **Coding Standards**
- Use **structured logging** (slog) not fmt.Print
- Handle **all errors explicitly** - never ignore them
- Use **context.Context** throughout for cancellation and timeouts
- Write **testable code** - separate business logic from AWS API calls
- Use **dependency injection** for AWS clients (easier to test)
- Follow **Go naming conventions** and add proper documentation

### **Implementation Approach**
1. **Start small** - get basic structure working first
2. **Research AWS SDK v2 patterns** for CloudWatch Logs, Config, and KMS
3. **Build incrementally** - one feature at a time
4. **Add comprehensive error handling** after core logic works
5. **Optimize for memory usage** - this runs on Lambda

### **What to Build**
- Lambda function that receives AWS Config compliance results
- Applies KMS encryption to unencrypted log groups  
- Sets retention policies on log groups without retention
- Handles multiple regions and batch processing
- Logs everything for debugging and audit trails

### **Testing Strategy**
- Create unit tests with mocked AWS services
- Add integration tests that can run against real AWS (optional)
- Include benchmarks for memory usage optimization

**Just start coding and ask questions when you get stuck. Make decisions and move forward.**