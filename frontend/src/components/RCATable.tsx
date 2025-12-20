import { RCAItem } from '../types';

interface RCATableProps {
  rcas: RCAItem[];
}

const severityStyles = {
  info: 'bg-blue-100 text-blue-800 border-blue-200',
  warning: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  critical: 'bg-red-100 text-red-800 border-red-200',
};

function RCATable({ rcas }: RCATableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full border-collapse border border-gray-300">
        <thead>
          <tr className="bg-gray-50">
            <th className="border border-gray-300 px-4 py-3 text-left text-sm font-semibold text-gray-700">
              Time
            </th>
            <th className="border border-gray-300 px-4 py-3 text-left text-sm font-semibold text-gray-700">
              Title
            </th>
            <th className="border border-gray-300 px-4 py-3 text-left text-sm font-semibold text-gray-700">
              Severity
            </th>
            <th className="border border-gray-300 px-4 py-3 text-left text-sm font-semibold text-gray-700">
              Edit
            </th>
          </tr>
        </thead>
        <tbody>
          {rcas.map((rca) => (
            <tr key={rca.id} className="hover:bg-gray-50">
              <td className="border border-gray-300 px-4 py-3 text-sm text-gray-700">
                {rca.time}
              </td>
              <td className="border border-gray-300 px-4 py-3 text-sm text-gray-700">
                {rca.title}
              </td>
              <td className="border border-gray-300 px-4 py-3 text-sm">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${severityStyles[rca.severity]}`}
                >
                  {rca.severity}
                </span>
              </td>
              <td className="border border-gray-300 px-4 py-3 text-sm">
                <button
                  className="px-3 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 transition-colors"
                  onClick={() => {
                    // Handle edit action
                    console.log('Edit RCA:', rca.id);
                  }}
                >
                  edit
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default RCATable;

