Plan of tags: 

```
SET user:123:session value ATAG session:123
SET user:123:prefs value ATAG session:123
SET user:123:cart value ATAG session:123

# user logs out
AEGIS.INVALIDATE TAG session:123
```