# SweetAlert2 Usage Guide

## 📦 Installation

```bash
yarn add sweetalert2
```

## 🚀 Usage

### Basic Usage

```typescript
import { showSuccess, showError, showWarning, showInfo, showConfirm } from '@/lib/sweetalert'

// Success alert
showSuccess('Success!', 'Data saved successfully')

// Error alert
showError('Error', 'Something went wrong')

// Warning alert
showWarning('Warning', 'Please review your input')

// Info alert
showInfo('Info', 'New update available')
```

### Confirmation Dialog

```typescript
import { showConfirm } from '@/lib/sweetalert'

const result = await showConfirm(
  'Delete Item',
  'Are you sure you want to delete this item?',
  'Delete',
  'Cancel'
)

if (result.isConfirmed) {
  // User confirmed
  await deleteItem()
  showSuccess('Deleted!', 'Item has been deleted')
} else {
  // User cancelled
  showInfo('Cancelled', 'Item was not deleted')
}
```

### Loading State

```typescript
import { showLoading, close, showSuccess } from '@/lib/sweetalert'

// Show loading
showLoading('Processing...')

// Simulate API call
setTimeout(() => {
  close()
  showSuccess('Done!', 'Process completed')
}, 2000)
```

### Custom Alert

```typescript
import { showCustom } from '@/lib/sweetalert'

showCustom({
  title: 'Custom Alert',
  html: '<div>Custom <b>HTML</b> content</div>',
  icon: 'info',
  confirmButtonText: 'Got it!',
  showCancelButton: true,
})
```

## 🎨 Pre-configured Theme

Alert sudah dikonfigurasi dengan project theme:
- **Confirm Button**: Primary blue (`#3b82f6`)
- **Cancel Button**: Gray (`#6b7280`)
- **Success**: Green
- **Error**: Red
- **Warning**: Orange
- **Info**: Cyan

## 📝 API Reference

| Function | Parameters | Returns | Description |
|----------|-----------|---------|-------------|
| `showSuccess` | `(title, text?)` | `Promise` | Success alert (auto-close 3s) |
| `showError` | `(title, text?)` | `Promise` | Error alert |
| `showWarning` | `(title, text?)` | `Promise` | Warning alert |
| `showInfo` | `(title, text?)` | `Promise` | Info alert |
| `showConfirm` | `(title, text?, confirmText?, cancelText?)` | `Promise` | Confirmation dialog |
| `showCustom` | `(options)` | `Promise` | Custom alert with SweetAlert options |
| `showLoading` | `(title)` | `Promise` | Loading state |
| `close` | `()` | `void` | Close current alert |
| `updateLoadingTitle` | `(title)` | `void` | Update loading title |

## 🔗 Documentation

- [SweetAlert2 Official Docs](https://sweetalert2.github.io/)
- [SweetAlert2 GitHub](https://github.com/sweetalert2/sweetalert2)
