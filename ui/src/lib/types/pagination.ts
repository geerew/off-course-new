import { array, number, object, type Output } from 'valibot';
import { CourseSchema } from './models';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type PaginationParams = {
	page: number;
	perPage: number;
	perPages: number[];
	totalItems: number;
	totalPages: number;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const PaginationSchema = object({
	page: number(),
	perPage: number(),
	totalItems: number(),
	totalPages: number(),
	items: array(CourseSchema)
});

export type Pagination = Output<typeof PaginationSchema>;
