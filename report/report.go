package report

import (
	"fmt"

	"github.com/JayantSeth/pfms/utils"
)

func GenBasicStructure() string {
	basic_structure := `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Report</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
      href="https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&display=swap"
      rel="stylesheet"
    />
  </head>
  <body>
    <div class="flex justify-center">
      <section class="border-2 mt-2 w-[95%] bg-blue-50">
        <h1 class="text-4xl text-center m-8 font-semibold">Ping Report</h1>
      </section>
    </div>

    <div class="flex justify-center gap-20 mt-10 flex-wrap mb-10">PLACEHOLDER</div>
  </body>
</html>
	`
	return basic_structure
}

func GenTable(n utils.Node) string {
	table := fmt.Sprintf(`
    <div>
      <div
        class="block max-w-sm p-6 bg-white border border-gray-200 rounded-lg shadow-sm hover:bg-gray-100 dark:bg-gray-800 dark:border-gray-700 dark:hover:bg-gray-700"
      >
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 text-center dark:text-white"
        >
          %s (%s)
        </h5>

        <div class="relative overflow-x-auto">
          <table
            class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400"
          >
            <thead
              class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400"
            >
              <tr>
                <th scope="col" class="px-6 py-3">IP</th>
                <th scope="col" class="px-6 py-3">Status</th>
              </tr>
            </thead>
            <tbody>
			ROWS_PLACEHOLDER
            </tbody>
          </table>
        </div>
      </div>
    </div>
	`, n.Name, n.IpAddress)
	return table
}


func GenRow(ip string, result bool) string {
	rString := `<td class="px-6 py-4">
	<span class="relative flex size-3">
	<span
		class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"
	></span>
	<span
		class="relative inline-flex size-3 rounded-full bg-green-500"
	></span>
	</span>
</td>
`
	if !result {
		rString = `
                <td class="px-6 py-4">
                  <span class="relative flex size-3">
                    <span
                      class="relative inline-flex size-3 rounded-full bg-red-500"
                    ></span>
                  </span>
                </td>
		`
	}
	row := fmt.Sprintf(`
<tr 
class="bg-white border-b dark:bg-gray-800 dark:border-gray-700 border-gray-200"
>
<th
	scope="row"
	class="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white"
>
	%s
</th>
%s
</tr>
	`, ip, rString)
	return row
}
