package req

import (
	"context"
	"sync"
)

// clientContext 为 Client 提供上下文能力，确保并发安全
type clientContext struct {
	mu      sync.RWMutex
	context context.Context
}

// WithContext 添加上下文键值对，使用读写锁保证并发安全
// 这种方式避免了 Clone 整个 Client 的开销
func (c *Client) WithContext(key, value any) *Client {
	if c.ctx == nil {
		c.ctx = &clientContext{
			context: context.Background(),
		}
	}

	c.ctx.mu.Lock()
	defer c.ctx.mu.Unlock()

	if c.ctx.context == nil {
		c.ctx.context = context.Background()
	}
	c.ctx.context = context.WithValue(c.ctx.context, key, value)
	return c
}

// GetContext 获取上下文值，使用读锁保证并发安全
func (c *Client) GetContext(key any) any {
	if c.ctx == nil {
		return context.Background().Value(key)
	}

	c.ctx.mu.RLock()
	defer c.ctx.mu.RUnlock()

	if c.ctx.context == nil {
		return nil
	}
	return c.ctx.context.Value(key)
}

// SetContext 设置基础上下文，使用写锁保证并发安全
func (c *Client) SetContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}

	if c.ctx == nil {
		c.ctx = &clientContext{}
	}

	c.ctx.mu.Lock()
	defer c.ctx.mu.Unlock()

	c.ctx.context = ctx
	return c
}

// Context 获取当前上下文，使用读锁保证并发安全
func (c *Client) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}

	c.ctx.mu.RLock()
	defer c.ctx.mu.RUnlock()

	if c.ctx.context == nil {
		return context.Background()
	}
	return c.ctx.context
}
