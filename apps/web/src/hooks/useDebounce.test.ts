import { act, renderHook } from '@testing-library/react';
import { useDebounce } from './useDebounce';

jest.useFakeTimers();

describe('useDebounce Hook', () => {
  afterEach(() => {
    jest.clearAllTimers();
  });

  it('should return initial value', () => {
    const { result } = renderHook(() => useDebounce('test', 500));
    expect(result.current).toBe('test');
  });

  it('should debounce value changes', () => {
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'initial', delay: 500 } },
    );

    expect(result.current).toBe('initial');

    rerender({ value: 'updated', delay: 500 });
    expect(result.current).toBe('initial'); // Not updated yet

    act(() => {
      jest.advanceTimersByTime(250);
    });
    expect(result.current).toBe('initial'); // Still not updated

    act(() => {
      jest.advanceTimersByTime(250);
    });
    expect(result.current).toBe('updated'); // Now updated
  });
});
