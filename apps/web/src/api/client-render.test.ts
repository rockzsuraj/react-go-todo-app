import { isRenderHibernateRateLimit } from './client';

describe('Render hibernation retry detection', () => {
  it('matches only Render hibernation 429 responses', () => {
    expect(isRenderHibernateRateLimit(429, 'hibernate-rate-limited')).toBe(true);
    expect(isRenderHibernateRateLimit(429, undefined)).toBe(false);
    expect(isRenderHibernateRateLimit(401, 'hibernate-rate-limited')).toBe(false);
  });
});
