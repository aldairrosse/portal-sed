import type { EvaluationProfile } from '$lib/types/evaluation';

export interface SimulatedUser {
	name: string;
	email: string;
}

export const PROFILE_USERS: Record<EvaluationProfile, SimulatedUser> = {
	colaborador: { name: 'María López García', email: 'maria.lopez@empresa.com' },
	jefe: { name: 'Carlos Rodríguez Pérez', email: 'carlos.rodriguez@empresa.com' },
	vendedor: { name: 'Ana Martínez Simón', email: 'ana.martinez@empresa.com' },
	'gerente-tienda': { name: 'Luis González Villa', email: 'luis.gonzalez@empresa.com' },
	divisional: { name: 'Patricia Fernández Díaz', email: 'patricia.fernandez@empresa.com' },
	regional: { name: 'Roberto Sánchez Ruiz', email: 'roberto.sanchez@empresa.com' },
	director: { name: 'Carmen Jiménez Castro', email: 'carmen.jimenez@empresa.com' },
	'director-general': { name: 'Carlos Mendoza', email: 'carlos.mendoza@empresa.com' },
	rh: { name: 'Laura Moreno Peña', email: 'laura.moreno@empresa.com' }
};