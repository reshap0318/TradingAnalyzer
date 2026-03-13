# TailwindCSS v4 Configuration

## 📋 Overview

Project ini menggunakan **TailwindCSS v4** dengan CSS-based configuration.

## 🎨 Custom Theme Colors

### Primary Colors
- `primary` - Main brand color (blue)
- `success` - Success states (green)
- `danger` - Error/danger states (red)
- `warning` - Warning states (orange)
- `info` - Information states (cyan)

### Trading Signal Colors
- `strong-buy` - Strong buy signal
- `buy` - Buy signal
- `wait` - Wait/hold signal
- `sell` - Sell signal
- `strong-sell` - Strong sell signal

## 📝 Usage Examples

### Utility Classes
```vue
<template>
  <!-- Primary colors -->
  <button class="bg-primary text-white hover:bg-primary-dark">
    Submit
  </button>
  
  <!-- Trading signals -->
  <div class="text-strong-buy font-bold">STRONG BUY</div>
  <div class="text-wait">WAIT</div>
  <div class="text-sell">SELL</div>
  
  <!-- Status colors -->
  <span class="bg-success-light text-success-dark px-2 py-1 rounded">
    Active
  </span>
  
  <span class="bg-danger-light text-danger-dark px-2 py-1 rounded">
    Error
  </span>
</template>
```

### Custom Container
```vue
<template>
  <div class="container">
    <!-- Content with auto margins and responsive padding -->
  </div>
</template>
```

## 🎯 Font Families

- `font-sans` - Inter (default)
- `font-mono` - JetBrains Mono / Fira Code (for code)

## 🌓 Dark Mode

Dark mode otomatis aktif berdasarkan system preference:
```css
@media (prefers-color-scheme: dark) {
  /* Dark theme styles */
}
```

## 📚 Documentation

- [TailwindCSS v4 Docs](https://tailwindcss.com/docs)
- [TailwindCSS v4 Migration Guide](https://tailwindcss.com/docs/upgrade-guide)
