# TextHighlighter Component Testing

This document describes the comprehensive test suite for the TextHighlighter component, ensuring robust behavior and preventing regressions in future development.

## Test Structure

### 1. Utility Functions Tests (`TextHighlighter.utils.test.js`)
Tests for pure functions extracted from the component:

#### Color Management
- `generateUniqueColor()` - Tests base color selection and HSL generation
- Color cycling and uniqueness validation
- Proper handling when all base colors are exhausted

#### Highlight Creation & Management
- `createHighlight()` - Tests highlight object creation with unique IDs
- `addHighlight()` - Tests adding highlights to arrays
- `removeHighlight()` - Tests highlight removal
- `updateHighlight()` - Tests highlight position updates

#### Word & Selection Logic
- `isWordInHighlight()` - Tests word-to-highlight containment
- `isWordInSelection()` - Tests selection state logic
- `findHighlightForWord()` - Tests highlight lookup by word index
- `checkOverlap()` - Tests overlap detection between highlights

#### Timestamp Conversion
- `calculateTimestamps()` - Tests word index to timestamp conversion
- `findWordByTimestamp()` - Tests timestamp to word index conversion
- Boundary handling and edge cases
- Round-trip conversion validation

#### Word Grouping
- `groupWordsAndHighlights()` - Tests grouping consecutive highlighted words
- Handles mixed regular words and highlight groups
- Proper space handling within highlighted regions

### 2. Integration Tests (`TextHighlighter.integration.test.js`)
Tests for complete workflows and component behavior:

#### Complete Workflows
- **Highlight Creation Workflow**: Selection → Creation → Color assignment → Grouping
- **Highlight Modification Workflow**: Update → Overlap checking → Removal → Re-grouping
- **Timestamp Conversion Workflow**: Word indices ↔ Timestamps round-trip
- **Selection Workflow**: Single word, multi-word, reversed selections

#### Complex Scenarios
- Multiple highlights with grouping
- Color management with recycling
- Large dataset performance testing
- Edge cases and error handling

#### Performance Tests
- Large dataset handling (1000+ words, 100+ highlights)
- Rapid timestamp lookups
- Grouping algorithm efficiency

## Test Coverage

### ✅ Covered Areas
1. **Pure Functions** (42 tests)
   - All utility functions with comprehensive edge cases
   - Input validation and error handling
   - Performance characteristics

2. **Integration Logic** (12 tests)
   - Complete user workflows
   - Cross-function interactions
   - Complex scenarios and edge cases

3. **Error Handling**
   - Empty inputs
   - Invalid indices
   - Malformed data
   - Out-of-bounds operations

4. **Performance**
   - Large dataset handling
   - Algorithm efficiency
   - Memory usage patterns

### 🔄 Areas for Future Testing
1. **Component Integration**
   - Svelte 5 component testing (requires better tooling setup)
   - DOM interaction testing
   - Event handling verification

2. **Accessibility**
   - Keyboard navigation
   - Screen reader compatibility
   - ARIA attributes

3. **Browser Compatibility**
   - Cross-browser event handling
   - CSS behavior consistency

## Running Tests

```bash
# Run all tests
pnpm test:run

# Run tests in watch mode
pnpm test

# Run tests with UI
pnpm test:ui

# Run specific test file
pnpm vitest run src/lib/components/TextHighlighter.utils.test.js
```

## Test Configuration

- **Framework**: Vitest with jsdom environment
- **Testing Library**: @testing-library/svelte for component testing
- **Mocking**: Vitest built-in mocking for controlled testing
- **Coverage**: Available via `vitest --coverage`

## Adding New Tests

When adding features to TextHighlighter:

1. **Add Pure Function Tests** to `TextHighlighter.utils.test.js`:
   - Test function in isolation
   - Include edge cases and error conditions
   - Verify immutability where applicable

2. **Add Integration Tests** to `TextHighlighter.integration.test.js`:
   - Test complete workflows
   - Verify interactions between functions
   - Include performance tests for complex operations

3. **Update This Documentation**:
   - Document new test scenarios
   - Update coverage information
   - Add examples for complex test cases

## Test Patterns

### Pure Function Testing
```javascript
describe('functionName', () => {
  it('should handle normal case', () => {
    expect(functionName(input)).toBe(expectedOutput);
  });
  
  it('should handle edge case', () => {
    expect(functionName(edgeInput)).toBe(expectedEdgeOutput);
  });
  
  it('should handle error case', () => {
    expect(() => functionName(invalidInput)).toThrow();
  });
});
```

### Integration Testing
```javascript
describe('workflow name', () => {
  it('should complete entire workflow', () => {
    // 1. Setup initial state
    // 2. Perform sequence of operations
    // 3. Verify final state
    // 4. Check side effects
  });
});
```

### Performance Testing
```javascript
it('should handle large datasets efficiently', () => {
  const largeDataset = createLargeDataset();
  const start = performance.now();
  const result = functionUnderTest(largeDataset);
  const duration = performance.now() - start;
  
  expect(duration).toBeLessThan(acceptableThreshold);
  expect(result).toMatchExpectedStructure();
});
```

## Debugging Failed Tests

1. **Check Test Output**: Read assertion messages carefully
2. **Use Console Logging**: Add temporary `console.log` statements
3. **Run Single Test**: Isolate failing test with `vitest run -t "test name"`
4. **Check Mocks**: Verify mock implementations match expectations
5. **Validate Test Data**: Ensure test fixtures represent real scenarios

## Maintenance

- **Review Tests Monthly**: Ensure tests still reflect current behavior
- **Update Test Data**: Keep sample data current with real usage patterns
- **Performance Benchmarks**: Monitor test execution times
- **Coverage Reports**: Regularly check for untested code paths