import { getCategories, getGoals } from '$lib/stores/goalsStore.svelte';
import type { GoalUnit } from '$lib/types/goal';

export const UNIT_OPTIONS: Array<{ value: GoalUnit; label: string }> = [
  { value: 'porcentaje', label: 'Porcentaje (%)' },
  { value: 'moneda', label: 'Moneda ($)' },
  { value: 'numero', label: 'Número' },
  { value: 'binario', label: 'Binario (Sí/No)' }
];

export function validateCategory(data: { name: string; description: string; weight: number; categoryId?: string }): string | null {
  if (!data.name.trim()) return 'El nombre es obligatorio.';
  if (!data.description.trim()) return 'La descripción es obligatoria.';
  if (data.weight < 0 || data.weight > 100) return 'El peso debe estar entre 0 y 100.';
  const trimmed = data.name.trim();
  const existing = getCategories();
  const duplicate = existing.find(c => c.name.toLowerCase() === trimmed.toLowerCase() && c.id !== data.categoryId);
  if (duplicate) return 'Ya existe una categoría con ese nombre.';
  return null;
}

export function validateGoal(data: {
  name: string;
  description: string;
  weight: number;
  targetValue: number;
  direction: 'ascendente' | 'descendente';
  baselineValue?: number;
  categoryId: string;
  goalId?: string;
}): string | null {
  if (!data.name.trim()) return 'El nombre es obligatorio.';
  if (!data.description.trim()) return 'La descripción es obligatoria.';
  if (data.weight < 0 || data.weight > 100) return 'El peso debe estar entre 0 y 100.';
  if (data.targetValue < 0) return 'El valor objetivo no puede ser negativo.';
  if (data.direction === 'descendente') {
    if (data.baselineValue === undefined || data.baselineValue === null || isNaN(data.baselineValue))
      return 'El valor inicial es obligatorio para metas descendentes.';
    if (data.baselineValue <= data.targetValue)
      return 'El valor inicial debe ser mayor al valor objetivo en metas descendentes.';
  }
  const trimmed = data.name.trim();
  const existing = getGoals().filter(g => g.categoryId === data.categoryId);
  const duplicate = existing.find(g => g.name.toLowerCase() === trimmed.toLowerCase() && g.id !== data.goalId);
  if (duplicate) return 'Ya existe una meta con ese nombre en esta categoría.';
  return null;
}
