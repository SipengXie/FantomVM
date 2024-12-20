use std::cmp::min;

use crate::types::{u256, FailStatus};

#[derive(Debug)]
pub struct Stack(Vec<u256>);

impl Stack {
    pub fn new(inner: &[u256]) -> Self {
        let len = min(inner.len(), 1024);
        let mut v = Vec::with_capacity(1024);
        v.extend_from_slice(&inner[..len]);
        Self(v)
    }

    pub fn as_slice(&self) -> &[u256] {
        self.0.as_slice()
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }

    pub fn push(&mut self, value: impl Into<u256>) -> Result<(), FailStatus> {
        if self.0.len() >= 1024 {
            return Err(FailStatus::StackOverflow);
        }
        #[cfg(feature = "unsafe-stack")]
        // SAFETY:
        // self.0 is initialized with capacity 1024 and never shrunk.
        unsafe {
            std::hint::assert_unchecked(self.0.capacity() == 1024);
        }
        self.0.push(value.into());
        Ok(())
    }

    pub fn swap_with_top(&mut self, nth: usize) -> Result<(), FailStatus> {
        self.check_underflow(nth + 1)?;

        #[cfg(not(feature = "unsafe-stack"))]
        {
            let len = self.0.len();
            self.0.swap(len - 1, len - 1 - nth);
        }
        #[cfg(feature = "unsafe-stack")]
        {
            let start = self.0.as_mut_ptr();
            // SAFETY:
            // This does not wrap and the whole range is valid.
            let top = unsafe { start.add(self.len() - 1) };
            // SAFETY:
            // This does not wrap and the whole range is valid.
            let nth = unsafe { top.sub(nth) };
            // SAFETY:
            // top and nth are valid pointers into the initialized part of the vector.
            unsafe {
                std::ptr::swap(top, nth);
            }
        }

        Ok(())
    }

    pub fn pop<const N: usize>(&mut self) -> Result<[u256; N], FailStatus> {
        self.check_underflow(N)?;

        let new_len = self.0.len() - N;
        let mut array = [u256::ZERO; N];
        array.copy_from_slice(&self.0[new_len..]);
        self.0.truncate(new_len);
        Ok(array)
    }

    pub fn peek(&self) -> Option<&u256> {
        self.0.last()
    }

    pub fn nth(&self, nth: usize) -> Result<u256, FailStatus> {
        self.check_underflow(nth + 1)?;
        #[cfg(not(feature = "unsafe-stack"))]
        return Ok(self.0[self.0.len() - 1 - nth]);
        #[cfg(feature = "unsafe-stack")]
        // SAFETY:
        // self.0.len() >= nth + 1 was checked in check_underflow.
        // Therefore self.0.len() - 1 - nth is in bounds.
        return Ok(*unsafe { self.0.get_unchecked(self.0.len() - 1 - nth) });
    }

    #[inline(always)]
    fn check_underflow(&self, min_len: usize) -> Result<(), FailStatus> {
        if self.0.len() < min_len {
            return Err(FailStatus::StackUnderflow);
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::types::{stack::Stack, u256, FailStatus};

    #[test]
    fn internals() {
        let stack = Stack::new(&[u256::ONE]);
        assert_eq!(stack.len(), 1);
        assert_eq!(stack.as_slice(), &[u256::ONE]);
    }

    #[test]
    fn push() {
        let mut stack = Stack::new(&[]);
        assert_eq!(stack.push(u256::MAX), Ok(()));
        assert_eq!(stack.as_slice(), [u256::MAX]);

        let mut stack = Stack::new(&[u256::ZERO; 1024]);
        assert_eq!(stack.push(u256::ZERO), Err(FailStatus::StackOverflow));
    }

    #[test]
    fn pop() {
        let mut stack = Stack::new(&[u256::MAX]);
        assert_eq!(stack.pop::<1>(), Ok([u256::MAX]));

        let mut stack = Stack::new(&[]);
        assert_eq!(stack.pop::<1>(), Err(FailStatus::StackUnderflow));

        let mut stack = Stack::new(&[u256::ONE, u256::MAX]);
        assert_eq!(stack.pop::<2>(), Ok([u256::ONE, u256::MAX]));

        let mut stack = Stack::new(&[u256::MAX]);
        assert_eq!(stack.pop::<2>(), Err(FailStatus::StackUnderflow));
    }

    #[test]
    fn nth() {
        let stack = Stack::new(&[u256::MAX, u256::ZERO]);
        assert_eq!(stack.nth(0), Ok(u256::ZERO));
        assert_eq!(stack.nth(1), Ok(u256::MAX));
        assert_eq!(stack.nth(2), Err(FailStatus::StackUnderflow));
    }

    #[test]
    fn swap_with_top() {
        let mut stack = Stack::new(&[u256::MAX, u256::ONE]);
        assert_eq!(stack.swap_with_top(0), Ok(()));
        assert_eq!(stack.as_slice(), &[u256::MAX, u256::ONE]);

        let mut stack = Stack::new(&[u256::MAX, u256::ONE]);
        assert_eq!(stack.swap_with_top(1), Ok(()));
        assert_eq!(stack.as_slice(), [u256::ONE, u256::MAX]);

        let mut stack = Stack::new(&[u256::MAX, u256::ONE]);
        assert_eq!(stack.swap_with_top(2), Err(FailStatus::StackUnderflow));
    }

    #[test]
    fn check_underflow() {
        let stack = Stack::new(&[]);
        assert_eq!(stack.check_underflow(0), Ok(()));
        let stack = Stack::new(&[u256::ZERO]);
        assert_eq!(stack.check_underflow(1), Ok(()));
        assert_eq!(stack.check_underflow(2), Err(FailStatus::StackUnderflow));
    }
}
