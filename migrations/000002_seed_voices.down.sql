-- voices の seed データを削除
DELETE FROM voices WHERE provider = 'google';

-- サンプルボイス音声を削除
DELETE FROM audios WHERE id IN (
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c01',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c02',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c03',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c04',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c05',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c06',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c07',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c08',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c09',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c10',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c11',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c12',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c13',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c14',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c15',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c16',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c17',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c18',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c19',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c20',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c21',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c22',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c23',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c24',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c25',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c26',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c27',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c28',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c29',
	'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c30'
);
