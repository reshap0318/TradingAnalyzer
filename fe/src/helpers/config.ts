/**
 * Config Key Helpers
 * Utility functions for formatting config keys and values
 */

/**
 * Convert config_key to human-readable label
 * @param key - Config key in UPPER_SNAKE_CASE format
 * @returns Formatted label in Title Case
 * 
 * @example
 * configKeyToLabel('MIN_CONFIDENCE') => 'Min Confidence'
 * configKeyToLabel('MAX_DAILY_TRADES') => 'Max Daily Trades'
 */
export function configKeyToLabel(key: string): string {
  return key
    .replace(/_/g, ' ')  // Replace underscore with space
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

/**
 * Get prefix/suffix for config value display
 * @param configKey - Config key identifier
 * @returns Object with prefix and/or suffix for value display
 * 
 * @example
 * getConfigValueSuffix('LEVERAGE') => { suffix: 'x' }
 * getConfigValueSuffix('MIN_CONFIDENCE') => { suffix: '%' }
 */
export function getConfigValueSuffix(configKey: string): { prefix?: string; suffix?: string } {
  const key = configKey.toUpperCase()
  
  // Percentage fields
  if (key.includes('PERCENT') || key.includes('CONFIDENCE') || key.includes('POSITION_SIZE')) {
    return { suffix: '%' }
  }
  
  // Leverage
  if (key.includes('LEVERAGE')) {
    return { suffix: 'x' }
  }
  
  // Time-related (hours)
  if (key.includes('HOURS')) {
    return { suffix: 'h' }
  }
  
  // Ratio fields (no suffix)
  if (key.includes('RATIO') || key.includes('TARGET') || key.includes('BUFFER')) {
    return { }
  }
  
  // Default: no prefix/suffix
  return { }
}

/**
 * Format config value for display with appropriate prefix/suffix
 * @param configKey - Config key identifier
 * @param value - Numeric value to format
 * @returns Formatted string with prefix/suffix
 * 
 * @example
 * formatConfigValue('LEVERAGE', 10) => '10x'
 * formatConfigValue('MIN_CONFIDENCE', 50) => '50%'
 */
export function formatConfigValue(configKey: string, value: number): string {
  const { prefix, suffix } = getConfigValueSuffix(configKey)
  
  // Special formatting for position size (convert decimal to percentage)
  if (configKey.toUpperCase().includes('POSITION_SIZE')) {
    value = value * 100
  }
  
  const formattedValue = Number.isInteger(value) ? value.toString() : value.toFixed(2)
  
  return `${prefix || ''}${formattedValue}${suffix || ''}`
}

/**
 * Check if config field should be boolean (radio button)
 * @param configKey - Config key identifier
 * @returns true if field is boolean
 */
export function isBooleanField(configKey: string): boolean {
  return configKey.toLowerCase().includes('is_')
}

/**
 * Check if config field should be decimal number
 * @param value - Config value string
 * @returns true if value is decimal
 */
export function isDecimalField(value: string): boolean {
  return value.includes('.')
}

/**
 * Get numeric value from config object
 * @param config - Config object with value property
 * @returns Numeric value or 0 if invalid
 */
export function getConfigNumericValue(config: any): number {
  const val = parseFloat(config.value)
  return isNaN(val) ? 0 : val
}

/**
 * Get input attributes for config field
 * @param config - Config object
 * @returns Input attributes object
 */
export function getInputAttrs(config: any): any {
  const configKey = config.config_key
  const value = config.value
  
  if (isBooleanField(configKey)) {
    return { type: 'radio' }
  }
  
  const isDecimal = isDecimalField(value)
  
  // Common attrs for number inputs
  const attrs: any = {
    type: 'number',
    step: isDecimal ? '0.01' : '1'
  }
  
  // Add min/max for specific fields
  if (configKey.includes('PERCENT') || configKey.includes('CONFIDENCE')) {
    attrs.min = 0
    attrs.max = 100
  } else if (configKey.includes('POSITION_SIZE')) {
    attrs.min = 0
    attrs.max = 1
    attrs.step = '0.01'
  } else if (configKey.includes('RATIO') || configKey.includes('TARGET')) {
    attrs.min = 0
    attrs.step = '0.1'
  } else if (configKey.includes('LEVERAGE')) {
    attrs.min = 1
  } else if (configKey.includes('HOURS') || configKey.includes('COUNT') || configKey.includes('TRADES')) {
    attrs.min = 1
  }
  
  return attrs
}
